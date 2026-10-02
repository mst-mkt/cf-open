package config

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

//go:embed load-typescript-config.mjs
var loaderScript string

const (
	loaderTimeout   = 30 * time.Second
	loaderWaitDelay = time.Second
)

type typeScriptConfig struct {
	AccountID  string            `json:"accountId"`
	Worker     *workerDefinition `json:"worker"`
	Containers []Container       `json:"containers"`
}

type workerDefinition struct {
	Name          string                   `json:"name"`
	Observability *ObservabilityConfig     `json:"observability"`
	Triggers      []workerTrigger          `json:"triggers"`
	Env           map[string]workerBinding `json:"env"`
	Exports       map[string]workerExport  `json:"exports"`
	TailConsumers []workerTailConsumer     `json:"tailConsumers"`
}

type workerTailConsumer struct {
	Worker    string `json:"worker"`
	Streaming bool   `json:"streaming"`
}

type workerReference string

func (w *workerReference) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		*w = workerReference(name)
		return nil
	}

	var definition struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &definition); err != nil {
		return err
	}

	*w = workerReference(definition.Name)
	return nil
}

type workerTrigger struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Schedule  string   `json:"schedule"`
	Pattern   string   `json:"pattern"`
	Zone      string   `json:"zone"`
	Addresses []string `json:"addresses"`
}

type workerExport struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type workerBinding struct {
	Type         string          `json:"type"`
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Jurisdiction string          `json:"jurisdiction"`
	Namespace    string          `json:"namespace"`
	TunnelID     string          `json:"tunnelId"`
	NetworkID    string          `json:"networkId"`
	StoreID      string          `json:"storeId"`
	SecretName   string          `json:"secretName"`
	ExportName   string          `json:"exportName"`
	Worker       workerReference `json:"worker"`
}

func loadTypeScriptConfig(configPath, mode string) (*Config, error) {
	node, err := exec.LookPath("node")
	if err != nil {
		return nil, errors.New("node not found in PATH; cloudflare.config.ts requires Node.js v22.18.0 or later")
	}

	output, err := runLoader(node, configPath, mode)
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("no config produced by %s", configPath)
	}

	config := &typeScriptConfig{}
	if err := json.Unmarshal(output, config); err != nil {
		return nil, fmt.Errorf("failed to parse the loaded config: %w", err)
	}

	cfg := toConfig(config.Worker, config.AccountID)
	cfg.Containers = config.Containers

	return cfg, nil
}

func runLoader(node, configPath, mode string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loaderTimeout)
	defer cancel()

	// The result comes back in a file rather than on stdout, which the config itself may write to.
	result, err := os.CreateTemp("", "cf-open-config-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temporary file: %w", err)
	}

	resultPath := result.Name()
	_ = result.Close()

	defer func() { _ = os.Remove(resultPath) }()

	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, node, loaderArgs(resultPath, configPath, mode)...)
	cmd.Stdin = strings.NewReader(loaderScript)
	cmd.Stderr = stderr
	cmd.WaitDelay = loaderWaitDelay

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to run node: %w", err)
	}

	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return nil, fmt.Errorf("timed out while evaluating %s", configPath)
	}

	if waitErr != nil {
		return nil, loaderError(configPath, stderr.String(), waitErr)
	}

	output, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the loaded config: %w", err)
	}

	return output, nil
}

func loaderArgs(resultPath, configPath, mode string) []string {
	args := []string{
		"--input-type=module",
		"--disable-warning=MODULE_TYPELESS_PACKAGE_JSON",
		"-",
		resultPath,
		configPath,
	}

	if mode == "" {
		return args
	}

	return append(args, mode)
}

func loaderError(configPath, stderr string, waitErr error) error {
	message := strings.TrimSpace(stderr)
	if message == "" {
		return fmt.Errorf("failed to evaluate %s: %w", configPath, waitErr)
	}

	return errors.New(message)
}

func toConfig(worker *workerDefinition, accountID string) *Config {
	if worker == nil {
		return &Config{AccountID: accountID}
	}

	env := worker.Env

	return &Config{
		Name:           worker.Name,
		AccountID:      accountID,
		Observability:  worker.Observability,
		Triggers:       cronTriggers(worker.Triggers),
		Routes:         fetchRoutes(worker.Triggers),
		Addresses:      emailAddresses(worker.Triggers),
		Queues:         queuesConfig(env, worker.Triggers),
		Workflows:      workflows(worker),
		DurableObjects: durableObjects(worker),
		Browser: firstBinding(env, "browser", func(binding string) BrowserConfig {
			return BrowserConfig{Binding: binding}
		}),
		AI: firstBinding(env, "ai", func(binding string) AIConfig {
			return AIConfig{Binding: binding}
		}),
		Stream: firstBinding(env, "stream", func(binding string) StreamConfig {
			return StreamConfig{Binding: binding}
		}),
		AISearch: collectBindings(env, "ai-search", func(binding string, instance workerBinding) (AISearchInstance, bool) {
			return AISearchInstance{Binding: binding, InstanceName: instance.Name}, instance.Name != ""
		}),
		AISearchNamespaces: collectBindings(env, "ai-search-namespace", func(binding string, namespace workerBinding) (AISearchNamespace, bool) {
			return AISearchNamespace{Binding: binding, Namespace: namespace.Namespace}, namespace.Namespace != ""
		}),
		AnalyticsEngineDatasets: collectBindings(env, "analytics-engine-dataset", func(binding string, dataset workerBinding) (AnalyticsEngineDataset, bool) {
			return AnalyticsEngineDataset{Binding: binding, Dataset: dataset.Name}, true
		}),
		SendEmail: collectBindings(env, "send-email", func(binding string, _ workerBinding) (SendEmail, bool) {
			return SendEmail{Name: binding}, true
		}),
		VPCServices: collectBindings(env, "vpc-service", func(binding string, service workerBinding) (VPCService, bool) {
			return VPCService{Binding: binding, ServiceID: service.ID}, service.ID != ""
		}),
		VPCNetworks: collectBindings(env, "vpc-network", func(binding string, network workerBinding) (VPCNetwork, bool) {
			return VPCNetwork{Binding: binding, TunnelID: network.TunnelID, NetworkID: network.NetworkID}, network.TunnelID != "" || network.NetworkID != ""
		}),
		Flagship: collectBindings(env, "flagship", func(binding string, flagship workerBinding) (Flagship, bool) {
			return Flagship{Binding: binding, AppID: flagship.ID}, true
		}),
		Services: collectBindings(env, "worker", func(binding string, service workerBinding) (Service, bool) {
			return Service{Binding: binding, Service: string(service.Worker)}, service.Worker != ""
		}),
		TailConsumers:          tailConsumers(worker.TailConsumers, false),
		StreamingTailConsumers: tailConsumers(worker.TailConsumers, true),
		R2Buckets: collectBindings(env, "r2", func(binding string, bucket workerBinding) (R2Bucket, bool) {
			return R2Bucket{Binding: binding, BucketName: bucket.Name, Jurisdiction: bucket.Jurisdiction}, bucket.Name != ""
		}),
		KVNamespaces: collectBindings(env, "kv", func(binding string, kv workerBinding) (KVNamespace, bool) {
			return KVNamespace{Binding: binding, ID: kv.ID}, kv.ID != ""
		}),
		D1Databases: collectBindings(env, "d1", func(binding string, db workerBinding) (D1Database, bool) {
			return D1Database{Binding: binding, DatabaseName: db.Name, DatabaseID: db.ID}, db.ID != ""
		}),
		Hyperdrive: collectBindings(env, "hyperdrive", func(binding string, hyperdrive workerBinding) (Hyperdrive, bool) {
			return Hyperdrive{Binding: binding, ID: hyperdrive.ID}, hyperdrive.ID != ""
		}),
		Pipelines: collectBindings(env, "pipeline", func(binding string, pipeline workerBinding) (Pipeline, bool) {
			return Pipeline{Binding: binding, Pipeline: pipeline.Name}, pipeline.Name != ""
		}),
		Vectorize: collectBindings(env, "vectorize", func(binding string, index workerBinding) (VectorizeIndex, bool) {
			return VectorizeIndex{Binding: binding, IndexName: index.Name}, index.Name != ""
		}),
		SecretsStoreSecrets: collectBindings(env, "secrets-store-secret", func(binding string, secret workerBinding) (SecretsStoreSecret, bool) {
			return SecretsStoreSecret{Binding: binding, StoreID: secret.StoreID, SecretName: secret.SecretName}, secret.StoreID != ""
		}),
		Images: firstBinding(env, "images", func(binding string) ImagesConfig {
			return ImagesConfig{Binding: binding}
		}),
	}
}

func bindingsOfKind(env map[string]workerBinding, kind string) iter.Seq2[string, workerBinding] {
	return func(yield func(string, workerBinding) bool) {
		for _, binding := range slices.Sorted(maps.Keys(env)) {
			def := env[binding]
			if def.Type != kind {
				continue
			}

			if !yield(binding, def) {
				return
			}
		}
	}
}

func collectBindings[T any](env map[string]workerBinding, kind string, convert func(string, workerBinding) (T, bool)) []T {
	var collected []T

	for binding, def := range bindingsOfKind(env, kind) {
		// convert reports whether the dashboard URL can be built from the binding.
		if converted, ok := convert(binding, def); ok {
			collected = append(collected, converted)
		}
	}

	return collected
}

func firstBinding[T any](env map[string]workerBinding, kind string, convert func(string) T) *T {
	for binding := range bindingsOfKind(env, kind) {
		converted := convert(binding)
		return &converted
	}

	return nil
}

func cronTriggers(triggers []workerTrigger) *TriggersConfig {
	var crons []string

	for _, trigger := range triggers {
		if trigger.Type == "scheduled" && trigger.Schedule != "" {
			crons = append(crons, trigger.Schedule)
		}
	}

	if len(crons) == 0 {
		return nil
	}

	return &TriggersConfig{Crons: crons}
}

func tailConsumers(consumers []workerTailConsumer, streaming bool) []TailConsumer {
	var collected []TailConsumer

	for _, consumer := range consumers {
		if consumer.Worker != "" && consumer.Streaming == streaming {
			collected = append(collected, TailConsumer{Service: consumer.Worker})
		}
	}

	return collected
}

func emailAddresses(triggers []workerTrigger) []string {
	var addresses []string

	for _, trigger := range triggers {
		if trigger.Type == "email" {
			addresses = append(addresses, trigger.Addresses...)
		}
	}

	return addresses
}

func fetchRoutes(triggers []workerTrigger) []Route {
	var routes []Route

	for _, trigger := range triggers {
		if trigger.Type != "fetch" || trigger.Pattern == "" {
			continue
		}

		route := Route{Pattern: trigger.Pattern}
		// cf treats a zone without a dot as a zone ID.
		if strings.Contains(trigger.Zone, ".") {
			route.ZoneName = trigger.Zone
		}

		routes = append(routes, route)
	}

	return routes
}

func queuesConfig(env map[string]workerBinding, triggers []workerTrigger) *QueuesConfig {
	producers := collectBindings(env, "queue", func(binding string, queue workerBinding) (QueueProducer, bool) {
		return QueueProducer{Binding: binding, Queue: queue.Name}, queue.Name != ""
	})

	var consumers []QueueConsumer
	for _, trigger := range triggers {
		if trigger.Type == "queue" && trigger.Name != "" {
			consumers = append(consumers, QueueConsumer{Queue: trigger.Name})
		}
	}

	if len(producers) == 0 && len(consumers) == 0 {
		return nil
	}

	return &QueuesConfig{Producers: producers, Consumers: consumers}
}

func workflows(worker *workerDefinition) []Workflow {
	var collected []Workflow

	add := func(workflow Workflow) {
		if workflow.Name == "" || slices.ContainsFunc(collected, func(w Workflow) bool { return w.Name == workflow.Name }) {
			return
		}

		collected = append(collected, workflow)
	}

	for binding, def := range bindingsOfKind(worker.Env, "workflow") {
		add(Workflow{Binding: binding, Name: def.Name, ClassName: def.ExportName})
	}

	for _, exportName := range slices.Sorted(maps.Keys(worker.Exports)) {
		if export := worker.Exports[exportName]; export.Type == "workflow" {
			add(Workflow{Name: export.Name, ClassName: exportName})
		}
	}

	return collected
}

func durableObjects(worker *workerDefinition) *DurableObjectsConfig {
	var bindings []DurableObjectBinding

	for binding, def := range bindingsOfKind(worker.Env, "durable-object") {
		bindings = append(bindings, DurableObjectBinding{Name: binding, ClassName: def.ExportName})
	}

	for _, exportName := range slices.Sorted(maps.Keys(worker.Exports)) {
		if worker.Exports[exportName].Type == "durable-object" {
			bindings = append(bindings, DurableObjectBinding{ClassName: exportName})
		}
	}

	if len(bindings) == 0 {
		return nil
	}

	return &DurableObjectsConfig{Bindings: bindings}
}
