package entries

import "github.com/uptrace/bun"

type PGAccount struct {
	bun.BaseModel `bun:"table:account"`

	Pkid                   string `bun:"pkid,pk"` // We make Pkid the primary key so that bun is happy.
	PublicKey              string `bun:"public_key"`
	Username               string `bun:"username"`
	Description            string `bun:"description"`
	ProfilePic             string `bun:"profile_pic"`
	CreatorBasisPoints     uint64 `bun:"creator_basis_points"`
	CoinWatermarkNanos     uint64 `bun:"coin_watermark_nanos"`
	MintingDisabled        bool   `bun:"minting_disabled"`
	DaoCoinMintingDisabled bool   `bun:"dao_coin_minting_disabled"`
	// TODO: bun seems to have a limit to the length of the string it can store in a column name.
	//DaoCoinTransferRestrictionStatus string                 `bun:"dao_coin_transfer_restriction_status"`
	ExtraData                 map[string]interface{} `bun:"extra_data"`
	CoinPriceDeSoNanos        uint64                 `bun:"coin_price_deso_nanos"`
	DeSoLockedNanos           uint64                 `bun:"deso_locked_nanos"`
	CCCoinsInCirculationNanos uint64                 `bun:"cc_coins_in_circulation_nanos"`
	//DAOCoinsInCirculationNanosHex uint64                 `bun:"dao_coins_in_circulation_nanos_hex"`
}
