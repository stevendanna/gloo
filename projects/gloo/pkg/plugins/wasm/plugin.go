package wasm

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/extend-envoy/pkg/cache"
	"github.com/solo-io/extend-envoy/pkg/defaults"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins"
	"github.com/solo-io/go-utils/protoutils"
)

const (
	FilterName = "envoy.filters.http.wasm"
)

type Plugin struct {
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Init(params plugins.InitParams) error {
	return nil
}

func (p *Plugin) plugin(pc *PluginSource) (*plugins.StagedHttpFilter, error) {

	cachedPlugin, err := p.ensurePluginInCache(pc)
	if err != nil {
		return nil, err
	}

	err = p.verifyConfiguration(cachedPlugin.Schema, pc.Config)
	if err != nil {
		return nil, err
	}

	filterCfg := &Wasm{
		Config: &PluginConfig{
			Name:          pc.Name,
			RootId:        pc.RootId,
			Configuration: pc.Config,
			VmConfig: &VmConfig{
				VmID:    "gloo-vm-id",
				Runtime: "envoy.wasm.runtime.wavm",
				Code: &AsyncDataSource{
					Remote: &RemoteDataSource{
						Sha256: cachedPlugin.Sha256,
						HttpUri: &HttpUri{
							Uri: "http://gloo/images/" + cachedPlugin.Sha256,
							// TODO: take from settings
							Cluster: "wasm-cache",
							Timeout: &types.Duration{
								Seconds: 5, // TODO: customize
							},
						},
					},
				},
			},
		},
	}

	jason, _ := json.Marshal(filterCfg)
	var strct types.Struct

	protoutils.UnmarshalBytes(jason, &strct)
	// TODO: allow customizing the stage
	stagedFilter, err := plugins.NewStagedFilterWithConfig(FilterName, &strct, plugins.DuringStage(plugins.AcceptedStage))
	if err != nil {
		return nil, err
	}

	return &stagedFilter, nil
}

var (
	imageCache cache.Cache
	once       sync.Once
)

func init() {
	imageCache = defaults.NewDefaultCache()
	go http.ListenAndServe(":9979", imageCache)
}

func (p *Plugin) ensurePluginInCache(pc *PluginSource) (*CachedPlugin, error) {

	digest, err := imageCache.Add(context.TODO(), pc.Image)
	if err != nil {
		return nil, err
	}
	return &CachedPlugin{
		Sha256: strings.TrimPrefix(string(digest), "sha256:"),
	}, nil
}

func (p *Plugin) verifyConfiguration(schema Schema, config string) error {
	// everything goes now-a-days
	return nil
}

func (p *Plugin) HttpFilters(params plugins.Params, l *v1.HttpListener) ([]plugins.StagedHttpFilter, error) {
	extensions := l.GetListenerPlugins().GetExtensions().GetConfigs()
	if extensions != nil {
		if config, ok := extensions["wasm"]; ok {
			jase, err := protoutils.MarshalBytes(config)
			if err != nil {
				return nil, err
			}
			pc := new(PluginSource)
			err = json.Unmarshal(jase, pc)
			if err != nil {
				return nil, err
			}

			stagedPlugin, err := p.plugin(pc)
			if err != nil {
				return nil, err
			}
			return []plugins.StagedHttpFilter{*stagedPlugin}, nil
		}
	}
	return nil, nil
}
