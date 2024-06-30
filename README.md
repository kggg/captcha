# captcha
A simple captcha in golang
#
example:
```
package main
import(
     "fmt"
     "github.com/kggg/captcha"
     )

func main(){
    capLen := 4 
    cap := captcha.NewCaptcha(capLen)
    //cap.SetStoreMode(captcha.NewRedisStore("127.0.0.1","passowrd","5379", 0, 300))
    cap.SetStoreMode(captcha.MemoryStoreMode)
    id, b64s, err := cap.GenerateCaptcha()
    if err != nil{
       fmt.Println(err)
    }
    fmt.Println("id: ", id)
    fmt.Println("base64 data: ", b64s)
   
}
```
