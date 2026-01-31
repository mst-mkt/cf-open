package cloudflare

import "testing"

func TestResource_Display(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resource Resource
		want     string
	}{
		{
			name: "説明がある場合は説明を返す",
			resource: Resource{
				Name:        "my-worker",
				Description: "Worker: my-worker",
			},
			want: "Worker: my-worker",
		},
		{
			name: "説明が空の場合は名前を返す",
			resource: Resource{
				Name:        "my-worker",
				Description: "",
			},
			want: "my-worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.resource.Display()
			if got != tt.want {
				t.Errorf("Resource.Display() = %q, want %q", got, tt.want)
			}
		})
	}
}
