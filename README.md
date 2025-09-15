# gowa-client

Cliente Go para a WhatsApp API MultiDevice (go-whatsapp-web-multidevice).

## Instalação

Requer Go 1.21+

```bash
go mod tidy
```

## Uso

```go
package main

import (
    "context"
    "fmt"
    "github.com/DantasBiao/gowa-client/pkg/gowa"
)

func main() {
    ctx := context.Background()

    cli, err := gowa.New(gowa.Config{
        BaseURL:  "http://localhost:3000",
        Username: "admin",
        Password: "admin",
    })
    if err != nil { panic(err) }

    login, err := cli.Login(ctx)
    if err != nil { panic(err) }
    fmt.Println("QR Link:", login.Results.QRLink)

    send, err := cli.SendMessage(ctx, "6289685028129@s.whatsapp.net", "Olá do Go!",
        gowa.WithForwarded(false),
        gowa.WithDisappearingDuration(3600),
    )
    if err != nil { panic(err) }
    fmt.Println("MessageID:", send.Results.MessageID)
}
```

## Convenções

- Autenticação via BasicAuth (username/password).
- Métodos de alto nível expõem parâmetros essenciais e `opts` funcionais para campos opcionais.
- `SendImageFile` usa multipart. `SendImageURL` usa JSON.

## Referência

- OpenAPI em `doc/openapi.yaml` (resumo local para desenvolvimento).
- API upstream: <https://github.com/aldinokemal/go-whatsapp-web-multidevice>
