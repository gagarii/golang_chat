package main

import (
  "log"
  "net/http"
  "text/template"
  "path/filepath"
  "sync"
  "flag"
  "oreilly/trace"
  "os"
  "github.com/stretchr/gomniauth"
  "github.com/stretchr/gomniauth/providers/facebook"
  "github.com/stretchr/gomniauth/providers/github"
  "github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
  once     sync.Once
  filename string
  templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  t.once.Do(func() {
    t.templ =
        template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
  })
  t.templ.Execute(w, r)
}

func main() {
  const clientId = "296075711013-6vl0mmn6r8alisotm9kb115l24b9k8gu.apps.googleusercontent.com"
  const privateAuthKey = "1mppUTzSj3f9XZWXk_kgBv2s"
  var addr = flag.String("addr", ":8080", "app-addr")
  flag.Parse() // フラグを開始します
  // Gomniauthのセットアップ
  gomniauth.SetSecurityKey("magari-golang-practice-key")
  gomniauth.WithProviders(
    facebook.New(clientId, privateAuthKey, "http://localhost:8080/auth/callback/facebook"),
    github.New(clientId, privateAuthKey, "http://localhost:8080/auth/callback/github"),
    google.New(clientId, privateAuthKey, "http://localhost:8080/auth/callback/google"),
  )
  r := newRoom()
  r.tracer = trace.New(os.Stdout)
  http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
  http.Handle("/login", &templateHandler{filename: "login.html"})
  http.HandleFunc("/auth/", loginHandler)
  http.Handle("/room", r)
  http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources/"))))
  // チャットルームを開始します
  go r.run()
  // Webサーバの起動
  log.Println("Webサーバーを開始します ポート：", *addr)
  if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
