# srpc
a simple and small pkg for bridging the multiplexed communication over stdio (or other similar interfaces), effective for a plugin based system.

# example plugin system
a full blown example can be found in the root

# loading the plugin for communication
```
cmd:=exec.Command("<your compiled plugin">)

bridge,errCh:=TwoWayBridgeWithACmd(cmd,nil)
// use call to do a req to plugin handler

resp,err:=bridge.Call(<>)
```