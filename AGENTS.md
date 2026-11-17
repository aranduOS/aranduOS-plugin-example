# AGENTS.md — plugin de parceiro

Este repositório é um plugin Go de parceiro. O programa de exemplo é
fidelidade (`example-loyalty`): desconto no `checkout.quote_adjust`, teto no
`order.pre_confirm`, pontos no `order.confirmed`.

A única dependência de plataforma é
[`github.com/mira-dev-tech/commerce-ext`](https://github.com/mira-dev-tech/commerce-ext).
As regras abaixo valem para este template e para o fork que o parceiro fizer
a partir dele. Cada regra tem a razão. Não as contorne para «ficar pronto».

## Regras

1. Não clone `mira-arandu-core` nem `mira-commerce-core`, e não copie este
   pacote para `extensions/` de lá. A razão: plugins escritos pela Mirá
   (`forms`, `care-record`, `condominio`) compilam dentro da API. O de
   parceiro é outro processo, noutro pod, e não herda a credencial da API.

2. Não use `MIRA_MANIFEST_PATH` e não descreva o core a arrancar este binário
   como filho. A razão: esse arranque dá ao plugin os privilégios do processo
   que serve os outros tenants. Aqui o `main` é `commerceext.Serve`. No
   cluster a única env é `COMMERCE_EXT_HOST`. O host disca
   `--commerce-ext-rpc`. O handshake `--commerce-ext-handshake` só prova o
   protocolo (`commerce-ext-ok`) e não publica nada.

3. Não peça DSN, `DATABASE_URL`, `kubectl` nem o mapa de tenants. A razão: o
   plugin não liga ao Postgres. O schema é `ext_<id>`, com hífen virado
   underscore, criado pelo host com um role sem superuser e sem `BYPASSRLS`.
   O contador em memória deste exemplo é didático.

4. Não mande preço nem chave de idempotência. Se for cobrar, `Usage.Add` leva
   o medidor, as vezes e a chave de negócio (`order_id`, não o id do evento).
   A razão: o preço é o da release que a empresa aceitou ao instalar, e a
   chave que evita a segunda cobrança é do host. Sem saldo o hook não corre.
   Timeout ou panic depois do débito estorna. Um deny de negócio não estorna.

5. Não chame `POST /platform/plugin-releases` com o token de parceiro. A
   sessão de parceiro leva 401 nessa rota. A razão: quem grava o digest é o
   workflow da Mirá, a partir da imagem que ele próprio construiu da tag.
   Um digest escrito neste repo não entra.

6. Não marque o pin `active` e não instale uma empresa que o registo não
   atribua a este parceiro. A razão: o install grava `installing`. `active`
   só depois de o host confirmar que o role respeita a RLS. Noutra empresa
   o host nem abre o RPC.

7. Hook tem 300ms e não faz I/O bloqueante. `LOYALTY_ORDER_CAP` (ou o código
   que o substituir no fork) é contrato: publicado, não se renomeia. Não
   logue segredo. Não ponha host de cluster, nome de cliente nem UUID real
   num repo que vá ser público.

## Publicar

1. PR. O CI deste repo (`go vet`, `go test ./...`, build, handshake) é a
   prova local. O `plugintest` substitui o core: cadeia, timeout, evento
   entregue duas vezes.

2. Tag `vX.Y.Z` no commit verde. Não se faz upload de binário.

3. O workflow da Mirá clona esse SHA, constrói, calcula o digest e grava a
   release com o `partner_id` ligado à org de GitHub. A primeira versão de
   um `plugin_id` novo espera review nesse environment. A seguinte, do mesmo
   id e do mesmo parceiro, não. A mesma versão não se regrava.

4. Ligar a uma empresa é um `workflow_dispatch` com o slug e o token de
   parceiro. O pin fica `installing` até a Mirá criar o role `ext_<id>` e
   activar. Suspender é disparo da Mirá e não reinicia a API.

O `vendor/` existe para o CI correr sem credencial de módulo. Não o apague.
Para subir o SDK: `go get github.com/mira-dev-tech/commerce-ext@vX.Y.Z`,
`go mod tidy`, `go mod vendor`, `go test ./...`.
