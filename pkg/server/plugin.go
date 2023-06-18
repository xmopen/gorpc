package server

import "gitee.com/zhenxinma/gorpc/pkg/errcode"

// PluginContainer 插件接口.
type PluginContainer interface {
	Add(plugin ...Plugin)
	DoRegister(name string, rcvr interface{}, metadata string) error
	CustomPlugDo(*Server) error
	Me() errcode.Error
}

// Plugin interface.
type Plugin interface{}

type (
	RegisterPlugin interface {
		Register(name string, server interface{}, metadata string) error
	}

	CustomPlugin interface {
		Do(*Server) error
	}
)

type pluginContainer struct {
	Plugs []Plugin
}

func (p *pluginContainer) Add(plugins ...Plugin) {
	for _, plugin := range plugins {
		p.Plugs = append(p.Plugs, plugin)
	}
}

func (p *pluginContainer) Me() errcode.Error {
	//TODO implement me
	panic("implement me")
}

// CustomPlugDo 自定义插件执行顺序应该在Server启动之前进行执行.
func (p *pluginContainer) CustomPlugDo(server *Server) error {
	for _, plug := range p.Plugs {
		if custom, ok := plug.(CustomPlugin); ok {
			// 转换成功.
			err := custom.Do(server)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DoRegister 执行所有RegisterPlugin的Register工作.
func (p *pluginContainer) DoRegister(name string, rcvr interface{}, metadata string) error {

	for _, plug := range p.Plugs {
		// 强转.
		if registerPlugin, ok := plug.(RegisterPlugin); ok {
			err := registerPlugin.Register(name, rcvr, metadata)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
