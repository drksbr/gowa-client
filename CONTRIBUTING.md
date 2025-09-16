# Guia de Contribuição para o gowa-client

Este documento orienta como contribuir com novos endpoints, evoluir o pacote e manter o versionamento e releases.

## 1. Como contribuir com novos endpoints

- Consulte o arquivo `doc/openapi.yaml` para entender os parâmetros e respostas dos endpoints.
- Implemente wrappers no pacote `pkg/gowa/client.go` seguindo o padrão dos métodos existentes:
  - Crie structs para parâmetros e respostas, se necessário.
  - Use o client HTTP já existente, com autenticação e tratamento de erros.
  - Adicione logs verbosos para facilitar debugging.
- Adicione exemplos de uso em `cmd/demo/main.go`.
- Teste localmente com `go run cmd/demo/main.go`.

## 2. Como abrir Pull Requests

- Crie um branch a partir de `main`.
- Faça commits pequenos e descritivos.
- Abra o PR com descrição clara do que foi alterado/adicionado.
- Aguarde revisão e aprovação.
- O CI irá validar build, vet e testes automaticamente.

## 3. Versionamento e Releases

- Após merge de PRs relevantes, crie uma tag semântica (ex: `v0.2.0`).
- O workflow de release irá publicar automaticamente no GitHub.
- Atualize o `CHANGELOG.md` com as principais mudanças.

## 4. Dúvidas e Sugestões

Abra uma issue no GitHub ou entre em contato pelo canal de suporte do projeto.

Obrigado por contribuir!

## 5. Exemplos Avançados e Boas Práticas

- Sempre utilize tipos explícitos para parâmetros e respostas, facilitando validação e documentação.
- Prefira contextos (`context.Context`) nos métodos para permitir cancelamento e timeout.
- Adicione testes unitários para novos métodos, se possível.
- Documente cada método novo com comentários GoDoc.
- Siga o padrão de logging já existente para facilitar troubleshooting.
- Para endpoints complexos, adicione exemplos detalhados em `cmd/demo/main.go`.

### Exemplo de Wrapper

```go
// Envia uma mensagem de texto para um número
func (c *Client) SendTextMessage(ctx context.Context, params SendTextParams) (*SendTextResponse, error) {
  // ...implementação...
}
```

### Exemplo de Teste

```go
func TestSendTextMessage(t *testing.T) {
  // ...setup e validação...
}
```

---

Siga estes exemplos para manter o padrão e facilitar a evolução do projeto.
