# aranduOS-plugin-example

**Template oficial de plugin do [aranduOS](https://aranduos.com).**
Um programa de fidelidade mínimo, comentado linha a linha, que você clona,
roda em 30 segundos e transforma no seu próprio plugin.

O que ele demonstra — as três superfícies que quase todo plugin real usa:

| Superfície | Hook/Evento | O que o exemplo faz |
|------------|-------------|---------------------|
| Hook que **ajusta** um resultado | `checkout.quote_adjust` | 5% de desconto quando o subtotal ≥ R$ 300 |
| Hook que **bloqueia** um fluxo | `order.pre_confirm` | Nega pedido acima de um teto (código estável `LOYALTY_ORDER_CAP`) |
| **Evento assíncrono** | `order.confirmed` | Credita pontos — idempotente pela chave de negócio (`order_id`) |

Exemplo concreto: um pedido de **R$ 879,90** sai do checkout por
**R$ 835,905** (desconto de R$ 43,995 aplicado pelo plugin, antes do pedido
ser persistido).

> A documentação completa do protocolo (catálogo de 18 hooks, 14 eventos,
> manifest, semânticas de falha) está no README do SDK:
> [`commerce-ext`](https://github.com/mira-dev-tech/commerce-ext).

---

## Comece em 30 segundos

```bash
git clone https://github.com/aranduOS/aranduOS-plugin-example
cd aranduOS-plugin-example
go test ./...        # funciona direto — o SDK está vendorado, zero setup
```

## Tour pelos arquivos (nesta ordem)

| Arquivo | O que você aprende lendo |
|---------|--------------------------|
| [`plugin.go`](plugin.go) | O plugin inteiro: `Meta/Init/Register/Shutdown`, os 2 hooks, o handler de evento, config via `Runtime` → env → default. **Cada decisão está comentada no lugar onde acontece.** |
| [`plugin_test.go`](plugin_test.go) | Testes unitários usando só a superfície do SDK — inclusive o teste que trava a paridade entre `Register` e o `manifest.yaml` |
| [`integration_test.go`](integration_test.go) | Integração local com o mini-host `plugintest`: cadeia, timeout, at-least-once — sem precisar do core |
| [`manifest.yaml`](manifest.yaml) | Como declarar capacidades (hooks/eventos) e compatibilidade de versão |
| [`cmd/example-loyalty/main.go`](cmd/example-loyalty/main.go) | O `main`: `commerceext.Serve`. Três linhas. Não é a API a arrancar o processo |
| [`AGENTS.md`](AGENTS.md) | Regras para quem (ou o agente de quem) for publicar este plugin, e a razão de cada uma |
| [`.github/workflows/ci.yml`](.github/workflows/ci.yml) | CI mínimo: vet + test + build + handshake |

## Como o binário sobe

Este template é o formato do parceiro. Não se copia para o monorepo e a API
não arranca o processo. O `main` chama `commerceext.Serve`. No cluster a
única env é `COMMERCE_EXT_HOST`: o host disca o RPC, o plugin não liga ao
Postgres e não vê o mapa de tenants.

```bash
go build -o bin/example-loyalty ./cmd/example-loyalty

# prova do protocolo, no CI — imprime commerce-ext-ok e sai
./bin/example-loyalty --commerce-ext-handshake
```

Se o processo morrer no meio de um hook, o resultado é `plugin_failure`. Não há manifest a remover nem restart da API para desligar: suspender a release tira o pin.

## Crie o seu plugin a partir deste template

1. **Fork/clone** com o nome `aranduOS-plugin-<seu-id>`.
2. **Renomeie a identidade**: `id` no `Meta()` e no `manifest.yaml`, o nome do
   pacote, o prefixo de env (`EXAMPLE_LOYALTY_*` → `SEU_PLUGIN_*`).
3. **Escolha hooks/eventos** no catálogo do SDK e **declare no manifest** — o
   teste `TestRegisterDeclaresManifestCapabilities` falha se divergirem.
4. **Escreva o handler mais burro possível primeiro**, com teste, e só depois
   sofistique. O exemplo mostra o padrão para cada tipo de handler.
5. **Rode `go test ./...` + o handshake** antes de qualquer deploy.

### As 5 regras de ouro (aprendidas em produção)

1. **Hook é síncrono com timeout de 300ms** — nada de I/O bloqueante; se
   estourar, o host degrada com `plugin_failure` (fail-closed em hooks de
   bloqueio).
2. **Handler de evento é idempotente pela CHAVE DE NEGÓCIO** (`order_id`),
   nunca pelo `Event.ID` — a entrega é at-least-once e o mesmo fato pode
   chegar em mais de um evento com IDs diferentes. Veja `onOrderConfirmed` no
   `plugin.go`.
3. **Código de bloqueio é contrato** — `LOYALTY_ORDER_CAP` é lido por front e
   suporte; publicou, não renomeia.
4. **Estado persistente vai para banco próprio do plugin** (schema
   `ext_<plugin>` no Postgres da plataforma). O contador em memória deste
   exemplo existe só para fins didáticos.
5. **Config via `Runtime.Config` → env → default; secrets nunca em log nem
   hardcode.**

## Testando o seu plugin — o que roda onde

Seja honesto com o seu pipeline: parte dos testes roda na sua máquina, parte
exige um ambiente da plataforma. O mapa completo:

| Etapa | Onde roda | Como |
|-------|-----------|------|
| **Unidade** — cada handler, config, idempotência | Sua máquina | `go test ./...` (padrões em [`plugin_test.go`](plugin_test.go)) |
| **Contrato go-plugin** — o binário fala o protocolo | Sua máquina | `go build -o bin/... ./cmd/... && ./bin/... --commerce-ext-handshake` → `commerce-ext-ok` |
| **Manifest** — capacidades declaradas = registradas | Sua máquina | teste `TestRegisterDeclaresManifestCapabilities` |
| **Integração local** — cadeia de prioridades, timeout de 300ms, semântica do `quote_adjust`, at-least-once de eventos | Sua máquina | mini-host [`plugintest`](https://github.com/mira-dev-tech/commerce-ext/tree/main/plugintest) — ver [`integration_test.go`](integration_test.go) |
| **Host real** — a release gravada a partir da tag, numa empresa do registo | Workflow da plataforma | A tag `vX.Y.Z` é o pedido. A plataforma constrói esse SHA e grava o digest. O pin fica `installing` até o role existir |
| **Aceite de negócio** — efeito num checkout de verdade | staging da plataforma | Pedido de teste num tenant de staging; valida desconto/bloqueio/pontos fim a fim |

Regra prática: **antes de pedir um slot em staging, unidade + integração
local (`plugintest`) + handshake + manifest têm que estar verdes** — são as
quatro coisas que você consegue provar sozinho.

## Colocando em produção

Tudo pelo GitHub. Não há painel neste caminho.

1. PR neste repo (ou no fork). `go test ./...` e o handshake verdes.
2. Tag `vX.Y.Z` no commit verde. O que se entrega é o fonte nesse SHA, não um binário.
3. O workflow da plataforma clona a tag, constrói a imagem e grava a release com o digest dele. O `partner_id` tem de estar ativo no registo dinâmico do Console. Um `plugin_id` de outro parceiro é recusado na gravação. A sessão de parceiro não consegue gravar a release.
4. Outro disparo liga a release a uma empresa que o registo diga ser deste parceiro. O pin fica `installing`.
5. A plataforma cria o role `ext_<id>` e activa. Sem o role, o pin não passa de `installing`. Suspender tira a release de todas as empresas e não reinicia a API.

As razões de cada passo estão no [`AGENTS.md`](AGENTS.md). Plugins da própria plataforma, compilados dentro da API, são outro caminho e não se fazem a partir deste template.

### Regras de versão

- Versione o plugin com semver e mantenha `compatibleCore` fiel (`^0.2.0`
  aceita toda a linha 0.2 do protocolo).
- Mudou código de bloqueio, hook usado ou formato de config? **Major** — e
  combine a migração antes do deploy.

## Atualizando a versão do SDK

O diretório `vendor/` congela o `commerce-ext` para o build funcionar sem
setup. Para subir de versão:

```bash
go get github.com/mira-dev-tech/commerce-ext@vX.Y.Z
go mod tidy && go mod vendor
go test ./...
```

## Troubleshooting

| Sintoma | Causa provável |
|---------|----------------|
| `--commerce-ext-handshake` não imprime `commerce-ext-ok` | O `main` não chama `commerceext.Serve(...)`, ou o binário é de outra arquitetura |
| Host recusa o plugin no load | `compatibleCore` do manifest incompatível com a linha do protocolo do core, ou hook usado sem estar declarado em `capabilities` |
| Hook "não dispara" | Confira o log do core (`ext: go-plugin RPC loaded id=...` / `ext: plugins loaded [...]`) e a prioridade — outro plugin pode ter bloqueado a cadeia antes |
| Checkout bloqueado com `plugin_failure` | Seu handler estourou 300ms, deu panic, ou o processo externo morreu — veja o log do core (`ext/rpc: hook ... failed`) |
| Evento processado 2×, pontos duplicados | Dedupe por `Event.ID` em vez da chave de negócio — releia a regra de ouro nº 2 |

## Sobre este repositório

Código de **exemplo didático**, fornecido como está. Forks são bem-vindos para
criar seus próprios plugins; a escrita aqui é restrita à equipe da plataforma para o
template permanecer um ponto de partida estável.

## Ecossistema

📚 **Documentação pública: [docs.mira-dev.tech](https://docs.mira-dev.tech)**

| Repo | O que é |
|---|---|
| [mira-commerce-sdk-js](https://github.com/mira-dev-tech/mira-commerce-sdk-js) | SDKs oficiais JS/TS — `client-sdk` (API tipada) e `checkout-sdk` (checkout headless) |
| [mira-commerce-storefront-guide](https://github.com/mira-dev-tech/mira-commerce-storefront-guide) | Guia oficial para construir um storefront (vitrine + checkout) sobre as APIs |
| [commerce-ext](https://github.com/mira-dev-tech/commerce-ext) | SDK público de plugins — Extension Protocol |
| [aranduOS-plugin-example](https://github.com/aranduOS/aranduOS-plugin-example) | Template de plugin — exemplo didático (hooks, eventos, manifest, go-plugin) |
| [aranduOS-plugin-payment](https://github.com/aranduOS/aranduOS-plugin-payment) | Template de plugin de pagamento e antifraude (roteamento, 3DS, velocity) |
| [aranduOS-plugin-shipping](https://github.com/aranduOS/aranduOS-plugin-shipping) | Template de plugin de logística/frete (cotação externa, cache, circuit breaker) |
