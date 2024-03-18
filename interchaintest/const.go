package interchaintest

import (
	sdkmath "cosmossdk.io/math"
	"github.com/liftedinit/manifest-ledger/x/manifest/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

const (
	votingPeriod     = "15s"
	maxDepositPeriod = "10s"
	Denom            = "umfx"

	accAddr  = "manifest1hj5fveer5cjtn4wd6wstzugjfdxzl0xp8ws9ct"
	acc2Addr = "manifest1efd63aw40lxf3n4mhf7dzhjkr453axurm6rp3z"

	userMnemonic       = "tuna develop gap truly crew canoe enlist slim stove scorpion clerk absurd better surprise moon fiction bean poem car air proud prevent unknown glue"
	cosmosGovModuleAcc = "manifest10d07y265gmmuvt4z0w9aw880jnsr700jmq3jzm"
	chainType          = "cosmos"
	chainName          = "manifest-ledger"
	chainID            = "manifest-2"
	chainRepository    = "ghcr.io/liftedinit/manifest-ledger"
	chainVersion       = "v0.0.1-alpha"
)

var (
	vals           = 1
	fullNodes      = 0
	DefaultGenesis = []cosmos.GenesisKV{
		// Governance
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", votingPeriod),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", maxDepositPeriod),
		// ABCI++
		cosmos.NewGenesisKV("consensus.params.abci.vote_extensions_enable_height", "1"),
		// TokenFactory
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_fee", nil),
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_gas_consume", "1"),
		// PoA
		cosmos.NewGenesisKV("app_state.poa.params.admins", []string{cosmosGovModuleAcc, accAddr}),
		// Mint - this is the only param the manifest module depends on from mint
		cosmos.NewGenesisKV("app_state.mint.params.blocks_per_year", "6311520"),
		// Manifest
		cosmos.NewGenesisKV("app_state.manifest.params.stake_holders", types.NewStakeHolders(types.NewStakeHolder(acc2Addr, 100_000_000))), // 100% of the inflation payout goes to them
		cosmos.NewGenesisKV("app_state.manifest.params.inflation.automatic_enabled", true),
		cosmos.NewGenesisKV("app_state.manifest.params.inflation.mint_denom", Denom),
		cosmos.NewGenesisKV("app_state.manifest.params.inflation.yearly_amount", "500000000000"), // in micro denom
	}

	LocalChainConfig = ibc.ChainConfig{
		Type:    chainType,
		Name:    chainName,
		ChainID: chainID,
		Images: []ibc.DockerImage{
			{
				Repository: chainRepository,
				Version:    chainVersion,
				UidGid:     "1025:1025",
			},
		},
		Bin:            "manifestd",
		Bech32Prefix:   "manifest",
		Denom:          Denom,
		GasPrices:      "0" + Denom,
		GasAdjustment:  1.3,
		TrustingPeriod: "508h",
		NoHostMount:    false,
		EncodingConfig: AppEncoding(),
		ModifyGenesis:  cosmos.ModifyGenesis(DefaultGenesis),
	}

	DefaultGenesisAmt = sdkmath.NewInt(10_000_000)
)
