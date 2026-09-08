package main
import (
 "fmt"
 "os"
 "path/filepath"
 "github.com/SuperMarioYL/mcpx/internal/clients"
 "github.com/SuperMarioYL/mcpx/internal/config"
)
type localClient struct { clients.Client; path string }
func (c localClient) ConfigPath() (string,bool,error) { return c.path,true,nil }
func main() {
 dir,err:=os.MkdirTemp("","mcpx-demo-"); if err!=nil {panic(err)}; defer os.RemoveAll(dir)
 spec:=clients.ServerSpec{Name:"demo",Command:"demo-server",Args:[]string{"--stdio"}}
 for _, id:=range []string{"claude-code","codex"} {
  adapter,_:=clients.ByID(id); path:=filepath.Join(dir,id+".config")
  initial:=[]byte(`{"theme":"keep-me","mcpServers":{}}`); if id=="codex" {initial=[]byte(`theme = "keep-me"`)}
  if err:=os.WriteFile(path,initial,0600);err!=nil {panic(err)}
  client:=localClient{adapter,path}
  first,err:=config.Merge(client,spec,false);if err!=nil {panic(err)}
  second,err:=config.Merge(client,spec,false);if err!=nil {panic(err)}
  items,err:=config.List(client);if err!=nil {panic(err)}
  fmt.Printf("%s: first_changed=%t backup=%t repeat_changed=%t servers=%d\n",id,first.Changed,first.BackupPath!="",second.Changed,len(items))
  result,err:=config.Remove(client,"demo",false);if err!=nil {panic(err)}
  fmt.Printf("%s: remove_changed=%t\n",id,result.Changed)
 }
}
