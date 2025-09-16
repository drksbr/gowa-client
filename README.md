# gowa-client

[![Go Reference](https://pkg.go.dev/badge/github.com/drksbr/gowa-client.svg)](https://pkg.go.dev/github.com/drksbr/gowa-client)

Cliente Go para a WhatsApp API MultiDevice (go-whatsapp-web-multidevice).

## Instalação

Requer Go 1.21+

```bash
go get github.com/drksbr/gowa-client@latest
go mod tidy
```

## Configuração

O cliente utiliza autenticação BasicAuth e permite configurar a URL base e timeout:

```go
cli, err := gowa.New(gowa.Config{
    BaseURL:  "http://localhost:3000", // ou seu endpoint
    Username: "admin",
    Password: "admin",
    Timeout:  20 * time.Second, // opcional
})
```

Ou via variáveis de ambiente:

```bash
export GOWA_BASE_URL="https://wa.provedorveloz.com.br/"
export GOWA_USER="admin"
export GOWA_PASS="sua_senha"
go run ./cmd/demo
```

## Exemplos de Uso

### Login QR (inicia sessão WhatsApp)

```go
login, err := cli.Login(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println("QR Link:", login.Results.QRLink)
```

### Enviar mensagem de texto

```go
send, err := cli.SendMessage(ctx, "558388572816@s.whatsapp.net", "Olá do Go!",
    gowa.WithForwarded(false),
    gowa.WithDisappearingDuration(3600),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("MessageID:", send.Results.MessageID)
```

### Enviar imagem (arquivo local)

```go
img, err := cli.SendImageFile(ctx, "558388572816@s.whatsapp.net", "Legenda", "./foto.jpg", false, false,
    gowa.WithDurationStr(3600),
)
```

### Enviar áudio

```go
audio, err := cli.SendAudio(ctx, gowa.SendAudioParams{
    Phone: "558388572816@s.whatsapp.net",
    AudioPath: "./audio.mp3",
    IsForwarded: false,
    Duration: 3600,
})
```

### Enviar contato

```go
contact, err := cli.SendContact(ctx, gowa.SendContactParams{
    Phone:        "558388572816@s.whatsapp.net",
    ContactName:  "Fulano",
    ContactPhone: "558388572816",
    IsForwarded:  false,
    Duration:     3600,
})
```

### Enviar localização

```go
loc, err := cli.SendLocation(ctx, gowa.SendLocationParams{
    Phone:     "558388572816@s.whatsapp.net",
    Latitude:  "-23.55052",
    Longitude: "-46.633308",
})
```

### Manipulação de mensagem

```go
// Revogar
_, err := cli.RevokeMessage(ctx, gowa.MessageActionParams{
    MessageID: "ID_DA_MSG",
    Phone:     "558388572816@s.whatsapp.net",
})
// Reagir
_, err := cli.ReactMessage(ctx, gowa.MessageActionParams{
    MessageID: "ID_DA_MSG",
    Phone:     "558388572816@s.whatsapp.net",
    Emoji:     "👍",
})
```

### Listar chats e mensagens

```go
chats, err := cli.ListChats(ctx, gowa.ListChatsParams{Limit: 10})
msgs, err := cli.GetChatMessages(ctx, "558388572816@s.whatsapp.net", gowa.GetChatMessagesParams{Limit: 20})
```

## Tratamento de erros

Todos os métodos retornam erro Go padrão. Se o erro for HTTP, a mensagem inclui o status e o corpo retornado.

## Dicas

- Sempre cheque erro antes de acessar campos da resposta.
- Use context com timeout para evitar travamentos.
- Os métodos aceitam structs de parâmetros para garantir tipagem e clareza.
- Para endpoints que aceitam arquivos, o caminho deve existir localmente.

## Principais tipos

- `gowa.Config`: configurações do cliente (BaseURL, Username, Password, Timeout)
- `gowa.Client`: instância principal
- `gowa.SendAudioParams`, `gowa.SendFileParams`, `gowa.SendContactParams`, etc: structs para payloads
- `gowa.MessageActionParams`: para manipulação de mensagens

## Referência de métodos

- `Login(ctx)`
- `SendMessage(ctx, phone, message, ...opts)`
- `SendImageFile(ctx, phone, caption, filePath, viewOnce, compress, ...opts)`
- `SendAudio(ctx, params)`
- `SendContact(ctx, params)`
- `SendLocation(ctx, params)`
- `RevokeMessage(ctx, params)`
- `ReactMessage(ctx, params)`
- `ListChats(ctx, params)`
- `GetChatMessages(ctx, chatJID, params)`

## Exemplo completo (demo)

Há um exemplo mínimo em `cmd/demo`. Para rodar localmente:

```bash
export GOWA_BASE_URL="http://localhost:3000"
export GOWA_USER="admin"
export GOWA_PASS="admin"
go run ./cmd/demo
```

## OpenAPI

Referência dos endpoints e payloads em `doc/openapi.yaml`.

## Upstream

- API original: <https://github.com/aldinokemal/go-whatsapp-web-multidevice>
