// Command example-loyalty é o processo do plugin de parceiro.
// Não é filho da API e não se carrega com MIRA_MANIFEST_PATH.
// No cluster a única env é COMMERCE_EXT_HOST; o host disca o RPC.
//
// Build:  go build -o bin/example-loyalty ./cmd/example-loyalty
// Prova:  ./bin/example-loyalty --commerce-ext-handshake   → imprime "commerce-ext-ok"
// Escuta: ./bin/example-loyalty --commerce-ext-rpc 0.0.0.0:8092
package main

import (
	commerceext "github.com/mira-dev-tech/commerce-ext"

	exampleloyalty "github.com/aranduOS/aranduOS-plugin-example"
)

func main() {
	commerceext.Serve(exampleloyalty.New())
}
