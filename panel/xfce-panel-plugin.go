package panel

// #cgo pkg-config: libxfce4panel-2.0
// #include <libxfce4panel/xfce-panel-plugin.h>
// #include "xfce-panel-plugin.go.h"
import "C"

import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

type Plugin struct {
	panelPlugin *C.XfcePanelPlugin
}

func GetPlugin() *Plugin {

	return xfceGoPlugin
}

func (p *Plugin) native() *C.XfcePanelPlugin {
	if p == nil || p.panelPlugin == nil {
		return nil
	}
	ptr := unsafe.Pointer(p.panelPlugin)
	return C.toXfcePanelPlugin(ptr)
}

func (p *Plugin) Object() *glib.Object {
	return glib.Take(
		unsafe.Pointer(
			p.panelPlugin,
		),
	)
}

func (p *Plugin) ConnectSig() {
	C.connectSig(p.native())
}
