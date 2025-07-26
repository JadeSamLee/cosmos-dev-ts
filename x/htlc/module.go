// x/htlc/module.go
package htlc

import (
    "github.com/cosmos/cosmos-sdk/types/module"
)

type AppModule struct{}

func NewAppModule() AppModule {
    return AppModule{}
}

func (AppModule) Name() string {
    return "htlc"
}

func (AppModule) RegisterInvariants(_ module.InvariantRegistry) {}

func (AppModule) Route() string {
    return "htlc"
}

func (AppModule) QuerierRoute() string {
    return "htlc"
}

func (AppModule) LegacyQuerierHandler(_ *module.LegacyQuerierHandler) {}

func (AppModule) RegisterServices(_ module.Configurator) {}

func (AppModule) InitGenesis(_ module.Generator) {}

func (AppModule) ExportGenesis(_ module.Generator) {}
