package consul

import (
	"context"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

func TestRegister(t *testing.T) {

	type args struct {
		ctx        context.Context
		serverName string
		server     []*register.ServiceInstance
	}

	test := []struct {
		name    string
		args    args
		want    []*register.ServiceInstance
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				serverName: "server-1",
				server: []*register.ServiceInstance{
					{
						ID:        "1",
						Name:      "server-1",
						Version:   "v0.0.1",
						Metadata:  nil,
						Endpoints: []string{"http://127.0.0.1:8000"},
					},
				},
			},
			want: []*register.ServiceInstance{
				{
					ID:        "1",
					Name:      "server-1",
					Version:   "v0.0.1",
					Metadata:  nil,
					Endpoints: []string{"http://127.0.0.1:8000"},
				},
			},
			wantErr: false,
		},
		{
			name: "registry new service replace old service",
			args: args{
				ctx:        context.Background(),
				serverName: "server-1",
				server: []*register.ServiceInstance{
					{
						ID:        "2",
						Name:      "server-1",
						Version:   "v0.0.1",
						Metadata:  nil,
						Endpoints: []string{"http://127.0.0.1:8000"},
					},
					{
						ID:        "2",
						Name:      "server-1",
						Version:   "v0.0.2",
						Metadata:  nil,
						Endpoints: []string{"http://127.0.0.1:8000"},
					},
				},
			},
			want: []*register.ServiceInstance{
				{
					ID:        "2",
					Name:      "server-1",
					Version:   "v0.0.2",
					Metadata:  nil,
					Endpoints: []string{"http://127.0.0.1:8000"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := api.NewClient(&api.Config{Address: "host.docker.internal:8500"})
			if err != nil {
				t.Fatalf("create consul client failed: %v", err)
			}
			r := NewRegistry(cli)
			for _, instance := range tt.args.server {
				err = r.Register(tt.args.ctx, instance)
				if err != nil {
					t.Error(err)
				}
			}

		})
	}
}
