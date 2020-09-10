package plugin

// #include <libxfce4panel/xfce-panel-plugin.h>
import "C"

import (
	"fmt"
	"log"
	"sort"
	"strconv"
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
	image, _ := gtk.ImageNewFromIconName("xfce4-lighter-plugin", gtk.ICON_SIZE_LARGE_TOOLBAR)
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

func switchForObj(widget *gtk.Switch, state bool, o interface{}) {
	switch obj := o.(type) {
	case huego.Light:
		if widget.GetActive() {
			obj.On()
		} else {
			obj.Off()
		}
	case huego.Group:
		if widget.GetActive() {
			obj.On()
		} else {
			obj.Off()
		}
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
	fillInPopupData(win)
	//fillInPopupRegister(win)

}

func fillInPopupRegister(win *gtk.Window) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
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

func fillInPopupData(win *gtk.Window) {

	groupBox := buildGroupBox()
	lightBox := buildLightBox()

	stack, _ := gtk.StackNew()
	stack.AddTitled(groupBox, "groups", "groups")
	stack.AddTitled(lightBox, "lights", "lights")
	stack.SetTransitionType(gtk.STACK_TRANSITION_TYPE_SLIDE_LEFT_RIGHT)
	stackSwitcher, _ := gtk.StackSwitcherNew()
	stackSwitcher.SetStack(stack)

	stackSwitcher.SetHAlign(gtk.ALIGN_CENTER)

	stackBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	stackBox.PackStart(stackSwitcher, true, true, 0)
	stackBox.PackStart(stack, true, true, 0)

	win.Add(stackBox)
	win.ShowAll()
}

func buildGroupBox() (box *gtk.Box) {
	groupIds, groupMap := getSortedGroups()
	box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 60)
	listBox, _ := gtk.ListBoxNew()
	listBox.SetSelectionMode(gtk.SELECTION_NONE)
	box.PackStart(listBox, true, true, 0)
	for _, groupID := range groupIds {
		group := groupMap[groupID]

		row, _ := gtk.ListBoxRowNew()
		vBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
		hBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 50)
		vBox.PackStart(hBox, true, true, 0)
		slider, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 1, 254, 1)
		slider.SetValue(float64(group.State.Bri))
		slider.SetDrawValue(false)

		vBox.PackStart(slider, true, true, 0)
		label, _ := gtk.LabelNew(group.Name)
		switchButton, _ := gtk.SwitchNew()
		switchButton.SetActive(group.IsOn())
		switchButton.SetVAlign(gtk.ALIGN_CENTER)
		switchButton.SetHAlign(gtk.ALIGN_END)
		hBox.PackStart(label, false, true, 0)
		hBox.PackStart(switchButton, true, true, 0)
		row.Add(vBox)
		listBox.Add(row)
		switchButton.Connect("state-set", switchForObj, group)
		if !group.IsOn() {
			slider.SetSensitive(false)
		}

		slider.Range.Connect("change-value", func(sliderRange *gtk.Scale, scrollType C.GtkScrollType, value float64, group *huego.Group) {
			var brightness uint8

			// Will do for each lamp as setting brightness with long scroll to group f#@ks things up

			// some boundaries
			// if value > 254 {
			// brightness = 254
			// } else if value < 1 {
			// 	brightness = 1
			// } else {
			// 	brightness = uint8(value)
			// }
			// fmt.Printf("brightness: %d\n", brightness)
			// if err := group.Bri(brightness); err != nil {
			// 	fmt.Println("Failed to update group light: ", err)
			// }

			sliderValue := sliderRange.GetValue()
			valueStep := value - sliderValue
			for _, light := range group.Lights {

				lightId, _ := strconv.Atoi(light)
				l, _ := bridge.GetLight(lightId)
				fmt.Printf("lamp: %s\n", l.Name)
				fmt.Printf("current brightness: %d\n", l.State.Bri)

				currentBri := float64(l.State.Bri)
				newBri := currentBri + valueStep

				if newBri > 254 {
					brightness = 254
				} else if newBri < 1 {
					brightness = 1
				} else {
					brightness = uint8(newBri)
				}
				fmt.Printf("new brightness: %d\n", brightness)
				l.Bri(brightness)
			}

		}, &group)

	}
	return
}

func buildLightBox() (box *gtk.Box) {
	box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 60)
	listBox, _ := gtk.ListBoxNew()
	listBox.SetSelectionMode(gtk.SELECTION_NONE)
	box.PackStart(listBox, true, true, 0)
	lightsIds, lights := getSortedLights()
	for _, lightID := range lightsIds {
		light := lights[lightID]
		row, _ := gtk.ListBoxRowNew()
		vBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
		hBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 50)
		vBox.PackStart(hBox, true, true, 0)
		slider, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 1, 254, 1)
		slider.SetValue(float64(light.State.Bri))
		slider.SetDrawValue(false)

		vBox.PackStart(slider, true, true, 0)
		label, _ := gtk.LabelNew(light.Name)
		switchButton, _ := gtk.SwitchNew()
		switchButton.SetActive(light.IsOn())
		switchButton.SetVAlign(gtk.ALIGN_CENTER)
		switchButton.SetHAlign(gtk.ALIGN_END)
		hBox.PackStart(label, false, true, 0)
		hBox.PackStart(switchButton, true, true, 0)
		row.Add(vBox)
		listBox.Add(row)
		switchButton.Connect("state-set", switchForObj, light)
		if !light.IsOn() {
			slider.SetSensitive(false)
		}

		slider.Range.Connect("change-value", func(sliderRange *gtk.Scale, scrollType C.GtkScrollType, value float64, light huego.Light) {
			var brightness uint8
			fmt.Printf("lamp: %s\n", light.Name)
			fmt.Printf("current brightness: %d\n", light.State.Bri)
			sliderValue := sliderRange.GetValue()
			valueStep := value - sliderValue
			currentBri := float64(light.State.Bri)
			newBri := currentBri + valueStep
			if newBri > 254 {
				brightness = 254
			} else if newBri < 1 {
				brightness = 1
			} else {
				brightness = uint8(newBri)
			}
			fmt.Printf("new brightness: %d\n", brightness)
			light.Bri(brightness)

		}, light)
	}
	return
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

func getSortedLights() (lightIds []int, lightMap map[int]huego.Light) {
	lightMap = make(map[int]huego.Light)
	lights, _ := bridge.GetLights()
	for _, light := range lights {
		lightMap[light.ID] = light
		lightIds = append(lightIds, light.ID)
	}
	sort.Ints(lightIds)
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
					fmt.Printf("Error updating %q: %s\n", light.Name, err)
				}
			}
		}
	}
}

//export AboutDialog
func AboutDialog() {
	about, _ := gtk.AboutDialogNew()
	about.AddCreditSection("Icons", []string{"May Canne"})
	about.SetProgramName("xfce4-lighther-plugin")
	about.SetAuthors([]string{"Michael Ketslah"})
	about.SetIconName("xfce4-lighter-plugin")
	about.ShowNow()

}

//export MenuDialog
func MenuDialog(plugin *C.XfcePanelPlugin) {
	C.xfce_panel_plugin_block_menu(xfceGoPlugin.native())
}
