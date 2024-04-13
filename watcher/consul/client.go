package consul

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/yanglunara/discovery/register"
)

var (
	_ register.Stopper = (*Client)(nil)
)

type Client struct {
	dc           register.DataCenter
	ctx          context.Context
	cancel       context.CancelFunc
	cli          *api.Client
	heartBeat    bool
	serviceCheck api.AgentServiceChecks

	enableHealthCheck bool
	maxTry            int

	deregisterCriticalServiceAfter time.Duration //它定义了一个服务在发生健康检查失败后，自动注销的时间
	timeout                        time.Duration
	healthCheckInterval            time.Duration

	entries register.Entries
}

func (c *Client) Service(ctx context.Context, service string, index uint64, passingOnly bool) ([]*register.ServiceInstance, uint64, error) {
	opts := &api.QueryOptions{
		WaitIndex: index,
		WaitTime:  time.Second * 55,
	}
	opts = opts.WithContext(ctx)
	if c.dc == register.MultiDataCenter {
		return c.entries.MultiDCService(ctx, &register.EntriesOption{
			Service:     service,
			Index:       index,
			PassingOnly: passingOnly,
			Opts:        opts,
		})
	}
	return c.entries.SingleDCEntries(ctx, &register.EntriesOption{
		Service:     service,
		PassingOnly: passingOnly,
		Opts:        opts,
		Index:       index,
	})

}

// Deregister 注销服务
func (c *Client) Deregister(_ context.Context, serviceID string) error {
	defer c.cancel()
	return c.cli.Agent().ServiceDeregister(serviceID)
}

// Register 注册服务
func (c *Client) Register(ctx context.Context, service *register.ServiceInstance) (err error) {
	address := make(map[string]api.ServiceAddress, len(service.Endpoints))
	checkAddress := make([]string, 0, len(service.Endpoints))
	for _, endpoint := range service.Endpoints {
		var raw *url.URL
		if raw, err = url.Parse(endpoint); err != nil {
			return
		}
		//端口号的范围就是 0 到 65535
		port, _ := strconv.ParseUint(raw.Port(), 10, 16)
		// 检查是否是合法的地址
		checkAddress = append(checkAddress,
			net.JoinHostPort(raw.Hostname(), strconv.Itoa(int(port))))
		address[raw.Scheme] = api.ServiceAddress{
			Address: endpoint,
			Port:    int(port),
		}
	}
	asr := &api.AgentServiceRegistration{
		ID:              service.ID,
		Name:            service.Name,
		Meta:            service.Metadata,
		Tags:            []string{0: fmt.Sprintf("version=%s", service.Version)},
		TaggedAddresses: address,
	}
	if len(checkAddress) > 0 {
		host, portRaw, _ := net.SplitHostPort(checkAddress[0])
		port, _ := strconv.ParseUint(portRaw, 10, 16)
		asr.Address = host
		asr.Port = int(port)
	}
	if c.enableHealthCheck {
		for _, adaddress := range checkAddress {
			asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
				TCP:                            adaddress,
				Interval:                       c.healthCheckInterval.String(),
				DeregisterCriticalServiceAfter: c.deregisterCriticalServiceAfter.String(),
				Timeout:                        c.timeout.String(),
			})
		}
		asr.Checks = append(asr.Checks, c.serviceCheck...)
	}
	// 开启心跳
	if c.heartBeat {
		newHealth := c.healthCheckInterval * 2
		asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
			CheckID:                        "service:%s" + service.ID,
			TTL:                            newHealth.String(),
			DeregisterCriticalServiceAfter: c.deregisterCriticalServiceAfter.String(),
		})
	}
	// 服务注册
	if err := c.cli.Agent().ServiceRegister(asr); err != nil {
		return err
	}
	if c.heartBeat {
		go c.startHearBeat(service.ID, asr)
	}
	return nil
}

// startHearBeat 开启心跳
func (c *Client) startHearBeat(serviceID string, asr *api.AgentServiceRegistration) {
	time.Sleep(time.Second * 1)
	if err := c.cli.Agent().UpdateTTL("service:"+serviceID, "pass", "pass"); err != nil {
		return
	}
	ticker := time.NewTicker(c.healthCheckInterval)
	defer ticker.Stop()
	deregister := func(agent *api.Agent, serviceID string) {
		_ = agent.ServiceDeregister(serviceID)
	}
	agent := c.cli.Agent()
	for {
		select {
		case <-c.ctx.Done():
			// 注销服务
			deregister(agent, serviceID)
			return
		case <-ticker.C:
			if errors.Is(c.ctx.Err(), context.Canceled) {
				deregister(c.cli.Agent(), serviceID)
				return
			}
			if err := agent.UpdateTTLOpts(
				"service:"+serviceID,
				"pass",
				"pass",
				new(api.QueryOptions).WithContext(c.ctx),
			); err != nil {
				if errors.Is(c.ctx.Err(), context.Canceled) ||
					errors.Is(err, context.DeadlineExceeded) {
					deregister(agent, serviceID)
					return
				}
				var (
					arrTry []int
				)
				for i := 0; i < c.maxTry; i++ {
					arrTry = append(arrTry, 1<<i)
					_ = agent.ServiceRegister(asr)
					// 重试间隔
					time.Sleep(time.Second * time.Duration(arrTry[i]))
				}
			}
		}
	}
}

func (c *Client) Close() error {
	c.cancel()
	return nil
}
