# lighter xfce4 panle plugin

![Alt Text](https://imgur.com/2bJ87qU.gif)

Simple xfce4 plugin that allows to turn on and off groups for hue lights 

TODO: Implement registration to hub

# Installation

Place icons:
 ```
 sudo cp  assets/xfce4-pulseaudio-plugin.png /usr/share/icons/hicolor/16x16/apps/
 ```

Update icons cache:
```
sudo gtk-update-icon-cache -f -t /usr/share/icons/hicolor
```
Compile:
```
go  build -v -buildmode c-shared -o liblighter.so 
```
Copy to panel:
```
sudo cp liblighter.so /usr/lib64/xfce4/panel/plugins/ && sudo cp liblighter.desktop /usr/share/xfce4/panel/plugins/
```
