package steam

type ListingItem struct {
	Success           bool                                   `json:"success"`
	PageSize          uint64                                 `json:"pagesize"`
	TotalCount        int                                    `json:"total_count"`
	Start             uint64                                 `json:"start"`
	NumActiveListings uint64                                 `json:"num_active_listings"`
	Assets            map[string]map[string]map[string]Asset `json:"assets"`
	Listings          []Listing                              `json:"listings"`
	ListingsOnHold    []Listing                              `json:"listings_on_hold"`
	ListingsToConfirm []Listing                              `json:"listings_to_confirm"`
	BuyOrders         []Listing                              `json:"buy_orders"`
}

type Asset struct {
	Currency                    uint64        `json:"currency"`
	AppID                       uint64        `json:"appid"`
	ContextID                   string        `json:"contextid"`
	ID                          string        `json:"id"`
	ClassID                     string        `json:"classid"`
	InstanceID                  string        `json:"instanceid"`
	Amount                      string        `json:"amount"`
	Status                      uint64        `json:"status"`
	OriginalAmount              string        `json:"original_amount"`
	UnownedID                   string        `json:"unowned_id"`
	UnownedContextID            string        `json:"unowned_contextid"`
	BackgroundColor             string        `json:"background_color"`
	IconURL                     string        `json:"icon_url"`
	IconURLLarge                string        `json:"icon_url_large"`
	Descriptions                []Description `json:"descriptions"`
	Tradable                    uint64        `json:"tradable"`
	Actions                     []Action      `json:"actions"`
	OwnerDescriptions           []Description `json:"owner_descriptions"`
	Name                        string        `json:"name"`
	NameColor                   string        `json:"name_color"`
	Type                        string        `json:"type"`
	MarketName                  string        `json:"market_name"`
	MarketHashName              string        `json:"market_hash_name"`
	Commodity                   uint64        `json:"commodity"`
	MarketTradableRestriction   int           `json:"market_tradable_restriction"`
	MarketMarketableRestriction uint64        `json:"market_marketable_restriction"`
	Marketable                  uint64        `json:"marketable"`
	AppIcon                     string        `json:"app_icon"`
	Owner                       uint64        `json:"owner"`
}

type Listing struct {
	ListingID                    string `json:"listingid"`
	TimeCreated                  uint64 `json:"time_created"`
	Asset                        Asset  `json:"asset"`
	SteamIDLister                string `json:"steamid_lister"`
	Price                        uint64 `json:"price"`
	OriginalPrice                uint64 `json:"original_price"`
	Fee                          uint64 `json:"fee"`
	CurrencyID                   string `json:"currencyid"`
	ConvertedPrice               uint64 `json:"converted_price"`
	ConvertedFee                 uint64 `json:"converted_fee"`
	ConvertedCurrencyID          string `json:"converted_currencyid"`
	Status                       uint64 `json:"status"`
	Active                       uint64 `json:"active"`
	SteamFee                     uint64 `json:"steam_fee"`
	ConvertedSteamFee            uint64 `json:"converted_steam_fee"`
	PublisherFee                 uint64 `json:"publisher_fee"`
	ConvertedPublisherFee        uint64 `json:"converted_publisher_fee"`
	PublisherFeePercent          string `json:"publisher_fee_percent"`
	PublisherFeeApp              uint64 `json:"publisher_fee_app"`
	CancelReason                 uint64 `json:"cancel_reason"`
	ItemExpired                  uint64 `json:"item_expired"`
	OriginalAmountListed         uint64 `json:"original_amount_listed"`
	OriginalPricePerUnit         uint64 `json:"original_price_per_unit"`
	FeePerUnit                   uint64 `json:"fee_per_unit"`
	SteamFeePerUnit              uint64 `json:"steam_fee_per_unit"`
	PublisherFeePerUnit          uint64 `json:"publisher_fee_per_unit"`
	ConvertedPricePerUnit        uint64 `json:"converted_price_per_unit"`
	ConvertedFeePerUnit          uint64 `json:"converted_fee_per_unit"`
	ConvertedSteamFeePerUnit     uint64 `json:"converted_steam_fee_per_unit"`
	ConvertedPublisherFeePerUnit uint64 `json:"converted_publisher_fee_per_unit"`
	TimeFinishHold               uint64 `json:"time_finish_hold"`
	TimeCreatedStr               string `json:"time_created_str"`
}

type Description struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Action struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

type SteamInventoryContext map[string]GameContext

type GameContext struct {
	AppID         uint64             `json:"appid"`
	Name          string             `json:"name"`
	Icon          string             `json:"icon"`
	Link          string             `json:"link"`
	AssetCount    uint64             `json:"asset_count"`
	InventoryLogo string             `json:"inventory_logo,omitempty"`
	TradePerms    string             `json:"trade_permissions"`
	LoadFailed    uint64             `json:"load_failed"`
	StoreVetted   string             `json:"store_vetted"`
	OwnerOnly     bool               `json:"owner_only"`
	RGContexts    map[string]Context `json:"rgContexts"`
}

type Context struct {
	AssetCount uint64 `json:"asset_count"`
	ID         string `json:"id"`
	Name       string `json:"name"`
}

type SearchData struct {
	Query              string `json:"query"`
	SearchDescriptions bool   `json:"search_descriptions"`
	TotalCount         uint64 `json:"total_count"`
	PageSize           uint64 `json:"pagesize"`
	Prefix             string `json:"prefix"`
	ClassPrefix        string `json:"class_prefix"`
}

type AssetDescription struct {
	AppID                       uint64 `json:"appid"`
	ClassID                     string `json:"classid"`
	InstanceID                  string `json:"instanceid"`
	Currency                    uint64 `json:"currency"`
	BackgroundColor             string `json:"background_color"`
	IconURL                     string `json:"icon_url"`
	IconURLLarge                string `json:"icon_url_large"`
	Tradable                    uint64 `json:"tradable"`
	Name                        string `json:"name"`
	NameColor                   string `json:"name_color"`
	Type                        string `json:"type"`
	MarketName                  string `json:"market_name"`
	MarketHashName              string `json:"market_hash_name"`
	Commodity                   uint64 `json:"commodity"`
	MarketTradableRestriction   int    `json:"market_tradable_restriction"`
	MarketMarketableRestriction uint64 `json:"market_marketable_restriction"`
	Marketable                  uint64 `json:"marketable"`
}

type MarketItem struct {
	Name             string           `json:"name"`
	HashName         string           `json:"hash_name"`
	SellListings     int              `json:"sell_listings"`
	SellPrice        int              `json:"sell_price"`
	SellPriceText    string           `json:"sell_price_text"`
	AppIcon          string           `json:"app_icon"`
	AppName          string           `json:"app_name"`
	AssetDescription AssetDescription `json:"asset_description"`
	SalePriceText    string           `json:"sale_price_text"`
}

type SteamMarketItems struct {
	Success    bool         `json:"success"`
	Start      int          `json:"start"`
	PageSize   int          `json:"pagesize"`
	TotalCount int          `json:"total_count"`
	SearchData SearchData   `json:"searchdata"`
	MarketItem []MarketItem `json:"results"`
}
