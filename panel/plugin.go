package panel

// #include <libxfce4panel/xfce-panel-plugin.h>
import "C"
import (
	"fmt"
	"log"
	"sort"
	"unsafe"

	"github.com/amimof/huego"
	"github.com/funkycode/lighter/hue"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var xfceGoPlugin *Plugin

var bridge hue.Bridge

// PluginBuild is exported function to init plugin
//export PluginBuild
func PluginBuild(plugin *C.XfcePanelPlugin) {
	xfceGoPlugin = &Plugin{}
	xfceGoPlugin.panelPlugin = plugin
	container := gtk.Container{gtk.Widget{glib.InitiallyUnowned{xfceGoPlugin.Object()}}}
	ebox, _ := gtk.EventBoxNew()
	ebox.Show()
	button, _ := gtk.ButtonNew()
	image, _ := gtk.ImageNewFromIconName("xfce4-lighter-plugin", gtk.ICON_SIZE_SMALL_TOOLBAR)
	button.SetImage(image)
	button.SetRelief(gtk.RELIEF_NONE)
	ebox.SetTooltipText("lighter")
	button.Show()
	container.Add(ebox)
	ebox.Add(button)
	button.AddEvents(int(gdk.SCROLL_MASK))
	button.Connect("clicked", lighterPopup, ebox.Object)
	//C.xfce_panel_plugin_add_action_widget(plugin, (*C.GtkWidget)(unsafe.Pointer(button.Native())))
	//C.xfce_panel_plugin_menu_show_configure(plugin)
	C.xfce_panel_plugin_menu_show_about(plugin)
	xfceGoPlugin.ConnectSig()
	xfceGoPlugin.Object().Connect("removed", func() {
		fmt.Println("menu removed")
	})
	scrollEventChan := make(chan gdk.ScrollDirection)
	xfceGoPlugin.Object().Connect("scroll-event", func(obj *glib.Object, e *gdk.Event) {
		scrollEvent := gdk.EventScrollNewFromEvent(e)
		scrollEventChan <- scrollEvent.Direction()
	})
	go updateOnScroll(scrollEventChan)
}

func switchForGroup(widget *gtk.Switch, state bool, group huego.Group) {
	if widget.GetActive() {
		group.On()
	} else {
		group.Off()
	}
}

func lighterPopup(obj *gtk.Button, parent *glib.Object) {

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetDecorated(false)
	win.Stick()
	win.SetSkipTaskbarHint(true)
	win.SetSkipPagerHint(true)
	var x, y C.gint
	C.xfce_panel_plugin_position_widget(xfceGoPlugin.native(),
		(*C.GtkWidget)(unsafe.Pointer(obj.Native())),
		(*C.GtkWidget)(unsafe.Pointer(parent.Native())),
		&x, &y)
	win.Move(int(x), int(y))
	win.Connect("destroy", func() {
		win.Destroy()
	})
	win.Connect("focus-out-event", func() {
		win.Destroy()
	})
	win.Connect("delete-event", func() {
		win.Destroy()
	})

	bridge.Bridge = hue.Connect()
	fillInPopupConnected(win)
	//fillInPopupRegister(win)

}

func fillInPopupRegister(win *gtk.Window) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	label, _ := gtk.LabelNew("Register bridge first")
	button, _ := gtk.ButtonNew()
	button.SetLabel("Register")

	button.Connect("clicked", func() {
		button.Destroy()
		label.SetLabel("Press button on bridge")
		sendNotification("lighter", "Press button on bridge")

	})
	box.PackStart(label, false, false, 0)
	box.PackStart(button, false, false, 0)
	win.Add(box)
	win.ShowAll()
}

func fillInPopupConnected(win *gtk.Window) {
	groupIds, groupMap := getSortedGroups()
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	for _, groupId := range groupIds {
		group := groupMap[groupId]
		innerBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
		label, _ := gtk.LabelNew(group.Name)
		switchButton, _ := gtk.SwitchNew()
		switchButton.SetActive(group.IsOn())
		innerBox.PackStart(label, true, true, 0)
		innerBox.PackStart(switchButton, false, false, 0)
		box.PackStart(innerBox, false, false, 0)
		switchButton.Connect("state-set", switchForGroup, group)
	}
	win.Add(box)
	win.ShowAll()
}

func sendNotification(title, body string) {
	appID := "org.lighter.notification"
	notification := glib.NotificationNew(title)
	notification.SetBody(body)
	app, _ := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	app.Connect("activate", func() {
		app.SendNotification(appID, notification)
	})
	app.Run(nil)
}

// Here we want to sort all groups so we do not get different order eahc click
func getSortedGroups() (groupIds []int, groupMap map[int]huego.Group) {
	groupMap = make(map[int]huego.Group)
	groups, _ := bridge.GetGroups()
	for _, group := range groups {
		if group.Type == "LightGroup" {
			continue
		}
		groupMap[group.ID] = group
		groupIds = append(groupIds, group.ID)
	}
	sort.Ints(groupIds)
	return
}

func updateOnScroll(scrollEventChan chan gdk.ScrollDirection) {
	for direction := range scrollEventChan {
		lights, err := bridge.GetLights()
		if err != nil {
			fmt.Println("Failed to get lights: ", err)
			continue
		}
		for _, light := range lights {
			if light.IsOn() {
				fmt.Printf("Current light %q: %d \n", light.Name, light.State.Bri)
				var brightness uint8
				if direction == gdk.SCROLL_DOWN {
					if light.State.Bri > 15 {
						brightness = light.State.Bri - 15
					} else {
						brightness = 1
					}
				} else if direction == gdk.SCROLL_UP {
					if light.State.Bri < 239 {
						brightness = light.State.Bri + 15
					} else {
						brightness = 254
					}
				} else {
					continue
				}
				fmt.Println("Setting new brightness: ", brightness)
				if err := light.Bri(brightness); err != nil {
					fmt.Println("Error updating %q: ", light.Name, err)
				}
			}
		}
	}
}

//export AboutDialog
func AboutDialog() {
	about, _ := gtk.AboutDialogNew()
	about.AddCreditSection("Icons", []string{"Maya Canne"})
	about.SetProgramName("xfce-lighter-plugin")
	about.SetAuthors([]string{"Michael Ketslah"})
	about.SetIconName("xfce4-lighter-plugin")
	about.ShowNow()

}

//export MenuDialog
func MenuDialog(plugin *C.XfcePanelPlugin) {
	C.xfce_panel_plugin_block_menu(xfceGoPlugin.native())
}
