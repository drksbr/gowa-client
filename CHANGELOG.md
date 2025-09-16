# Changelog

Todas as mudanças notáveis neste projeto serão documentadas aqui.

## v0.1.0 — 2025-09-16

- Primeira versão pública da biblioteca importável
- `module`: `github.com/drksbr/gowa-client`
- Pacote: `github.com/drksbr/gowa-client/pkg/gowa`
- Cliente HTTP com:
  - Basic Auth, timeout configurável
  - Retries com `go-retryablehttp`
  - Helpers JSON e multipart
- Métodos implementados (parciais do OpenAPI):
  - App: `Login`, `LoginWithCode`, `Logout`, `Reconnect`
  - Send: `SendMessage`, `SendImageFile`, `SendImageURL`, `SendAudio`, `SendFile`, `SendVideo`, `SendContact`, `SendLink`, `SendLocation`, `SendPoll`, `SendPresence`, `SendChatPresence`
  - Message: `RevokeMessage`, `DeleteMessage`, `ReactMessage`, `UpdateMessage`, `ReadMessage`, `StarMessage`, `UnstarMessage`
  - Chat: `ListChats`, `GetChatMessages`
- Exemplo: `cmd/demo` com logs verbosos
