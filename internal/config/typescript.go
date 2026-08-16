package config

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Worker   *workerDefinition   `json:"worker"`
	Settings *settingsDefinition `json:"settings"`
}

type workerDefinition struct {
	Name          string                   `json:"name"`
	Observability *ObservabilityConfig     `json:"observability"`
	Triggers      []workerTrigger          `json:"triggers"`
	Env           map[string]workerBinding `json:"env"`
}

type workerTrigger struct {
	Type     string `json:"type"`
	Schedule string `json:"schedule"`
}

type workerBinding struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	StoreID    string `json:"storeId"`
	SecretName string `json:"secretName"`
}

type settingsDefinition struct {
	AccountID string `json:"accountId"`
}

func (settings *settingsDefinition) accountID() string {
	if settings == nil {
		return ""
	}

	return settings.AccountID
}

func loadTypeScriptConfig(configPath string) (*WranglerConfig, error) {
	node, err := exec.LookPath("node")
	if err != nil {
		return nil, errors.New("node not found in PATH; cloudflare.config.ts requires Node.js v22.18.0 or later")
	}

	output, err := runLoader(node, configPath, os.Getenv("CLOUDFLARE_ENV"))
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

	if config.Worker == nil {
		return nil, fmt.Errorf("no worker found in %s", configPath)
	}

	return toWranglerConfig(config.Worker, config.Settings), nil
}

func runLoader(node, configPath, mode string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), loaderTimeout)
	defer cancel()

	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open a pipe: %w", err)
	}
	defer func() { _ = reader.Close() }()

	// A grandchild can inherit fd 3 and outlive node, so the read must not wait for it.
	if err := reader.SetReadDeadline(time.Now().Add(loaderTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set a read deadline: %w", err)
	}

	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, node, loaderArgs(configPath, mode)...)
	cmd.Stdin = strings.NewReader(loaderScript)
	cmd.Stderr = stderr
	// The result comes back on fd 3 rather than stdout, which the config itself may write to.
	cmd.ExtraFiles = []*os.File{writer}
	cmd.WaitDelay = loaderWaitDelay

	startErr := cmd.Start()
	_ = writer.Close()

	if startErr != nil {
		return nil, fmt.Errorf("failed to run node: %w", startErr)
	}

	output, readErr := io.ReadAll(reader)
	waitErr := cmd.Wait()

	if ctx.Err() != nil || errors.Is(readErr, os.ErrDeadlineExceeded) {
		return nil, fmt.Errorf("timed out while evaluating %s", configPath)
	}

	if waitErr != nil {
		return nil, loaderError(configPath, stderr.String(), waitErr)
	}

	if readErr != nil {
		return nil, fmt.Errorf("failed to read the loaded config: %w", readErr)
	}

	return output, nil
}

func loaderArgs(configPath, mode string) []string {
	args := []string{
		"--input-type=module",
		"--disable-warning=MODULE_TYPELESS_PACKAGE_JSON",
		"-",
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

func toWranglerConfig(worker *workerDefinition, settings *settingsDefinition) *WranglerConfig {
	env := worker.Env

	return &WranglerConfig{
		Name:          worker.Name,
		AccountID:     settings.accountID(),
		Observability: worker.Observability,
		Triggers:      cronTriggers(worker.Triggers),
		Queues:        queuesConfig(env),
		Browser: firstBinding(env, "browser", func(binding string) BrowserConfig {
			return BrowserConfig{Binding: binding}
		}),
		VPCServices: collectBindings(env, "vpc-service", func(binding string, service workerBinding) (VPCService, bool) {
			return VPCService{Binding: binding, ServiceID: service.ID}, service.ID != ""
		}),
		R2Buckets: collectBindings(env, "r2", func(binding string, bucket workerBinding) (R2Bucket, bool) {
			return R2Bucket{Binding: binding, BucketName: bucket.Name}, bucket.Name != ""
		}),
		KVNamespaces: collectBindings(env, "kv", func(binding string, kv workerBinding) (KVNamespace, bool) {
			return KVNamespace{Binding: binding, ID: kv.ID}, kv.ID != ""
		}),
		D1Databases: collectBindings(env, "d1", func(binding string, db workerBinding) (D1Database, bool) {
			return D1Database{Binding: binding, DatabaseName: db.Name, DatabaseID: db.ID}, db.ID != ""
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

func queuesConfig(env map[string]workerBinding) *QueuesConfig {
	producers := collectBindings(env, "queue", func(binding string, queue workerBinding) (QueueProducer, bool) {
		return QueueProducer{Binding: binding, Queue: queue.Name}, queue.Name != ""
	})

	if len(producers) == 0 {
		return nil
	}

	return &QueuesConfig{Producers: producers}
}
