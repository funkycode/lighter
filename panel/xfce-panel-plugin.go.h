extern void PluginBuild(XfcePanelPlugin *plugin);
extern void AboutDialog();
extern void MenuDialog(XfcePanelPlugin *plugin);

static XfcePanelPlugin *
toXfcePanelPlugin(void *p)
{
    return XFCE_PANEL_PLUGIN(p);
}

static XfcePanelPluginClass *
toXfcePanelPluginClass(void *p)
{
    return (XFCE_PANEL_PLUGIN_CLASS(p));
}

static void *
connectSig(XfcePanelPlugin *plugin)
{
    g_signal_connect(G_OBJECT(plugin), "about",
                     G_CALLBACK(AboutDialog),
                     NULL);
    g_signal_connect(G_OBJECT(plugin), "configure-plugin",
                     G_CALLBACK(MenuDialog),
                     NULL);
}

XFCE_PANEL_PLUGIN_REGISTER(PluginBuild)
