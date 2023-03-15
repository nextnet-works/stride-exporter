package main

import (
	"context"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	minttypes "github.com/Stride-Labs/stride/v6/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"

	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func ParamsHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	paramsMaxValidatorsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_max_validators",
			Help:        "Active set length",
			ConstLabels: ConstLabels,
		},
	)

	paramsUnbondingTimeGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_unbonding_time",
			Help:        "Unbonding time, in seconds",
			ConstLabels: ConstLabels,
		},
	)

	paramsBlocksPerYearGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_blocks_per_year",
			Help:        "Block per year",
			ConstLabels: ConstLabels,
		},
	)

	paramsAPYGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "stride_apy",
			Help:        "APY",
			ConstLabels: ConstLabels,
		},
	)

	paramsInflationGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "stride_inflation",
			Help:        "inflation rate",
			ConstLabels: ConstLabels,
		},
	)

	paramsInflationMinGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_inflation_min",
			Help:        "Min inflation",
			ConstLabels: ConstLabels,
		},
	)

	paramsInflationMaxGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_inflation_max",
			Help:        "Max inflation",
			ConstLabels: ConstLabels,
		},
	)

	paramsInflationRateChangeGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_inflation_rate_change",
			Help:        "Inflation rate change",
			ConstLabels: ConstLabels,
		},
	)

	paramsDowntailJailDurationGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_downtail_jail_duration",
			Help:        "Downtime jail duration, in seconds",
			ConstLabels: ConstLabels,
		},
	)

	paramsMinSignedPerWindowGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_min_signed_per_window",
			Help:        "Minimal amount of blocks to sign per window to avoid slashing",
			ConstLabels: ConstLabels,
		},
	)

	paramsSignedBlocksWindowGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_signed_blocks_window",
			Help:        "Signed blocks window",
			ConstLabels: ConstLabels,
		},
	)

	paramsSlashFractionDoubleSign := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_slash_fraction_double_sign",
			Help:        "% of tokens to be slashed if double signing",
			ConstLabels: ConstLabels,
		},
	)

	paramsSlashFractionDowntime := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_slash_fraction_downtime",
			Help:        "% of tokens to be slashed if downtime",
			ConstLabels: ConstLabels,
		},
	)

	paramsBaseProposerRewardGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_base_proposer_reward",
			Help:        "Base proposer reward",
			ConstLabels: ConstLabels,
		},
	)

	paramsBonusProposerRewardGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_bonus_proposer_reward",
			Help:        "Bonus proposer reward",
			ConstLabels: ConstLabels,
		},
	)
	paramsCommunityTaxGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_params_community_tax",
			Help:        "Community tax",
			ConstLabels: ConstLabels,
		},
	)
	paramsRedemptionRateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "stride_params_redemption_rate",
			Help:        "Redemption rate",
			ConstLabels: ConstLabels,
		},
		[]string{"chain"},
	)
	paramsStakedAmountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "stride_staked_amount",
			Help:        "Staked amount for each asset in Stride chain",
			ConstLabels: ConstLabels,
		},
		[]string{"chain"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(paramsMaxValidatorsGauge)
	registry.MustRegister(paramsUnbondingTimeGauge)
	registry.MustRegister(paramsBlocksPerYearGauge)
	registry.MustRegister(paramsInflationMinGauge)
	registry.MustRegister(paramsInflationMaxGauge)
	registry.MustRegister(paramsInflationRateChangeGauge)
	registry.MustRegister(paramsDowntailJailDurationGauge)
	registry.MustRegister(paramsMinSignedPerWindowGauge)
	registry.MustRegister(paramsSignedBlocksWindowGauge)
	registry.MustRegister(paramsSlashFractionDoubleSign)
	registry.MustRegister(paramsSlashFractionDowntime)
	registry.MustRegister(paramsBaseProposerRewardGauge)
	registry.MustRegister(paramsBonusProposerRewardGauge)
	registry.MustRegister(paramsCommunityTaxGauge)
	registry.MustRegister(paramsRedemptionRateGauge)
	registry.MustRegister(paramsStakedAmountGauge)
	registry.MustRegister(paramsAPYGauge)
	registry.MustRegister(paramsInflationGauge)

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying global staking params")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		paramsResponse, err := stakingClient.Params(
			context.Background(),
			&stakingtypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get global staking params")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying global staking params")

		paramsMaxValidatorsGauge.Set(float64(paramsResponse.Params.MaxValidators))
		paramsUnbondingTimeGauge.Set(paramsResponse.Params.UnbondingTime.Seconds())
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying global mint params")
		// queryStart := time.Now()

		mintClient := minttypes.NewQueryClient(grpcConn)
		paramsResponse, err := mintClient.Params(
			context.Background(),
			&minttypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get global mint params")
			return
		}

		epoch_provision := paramsResponse.Params.GenesisEpochProvisions
		staking := paramsResponse.Params.DistributionProportions.Staking
		reduction_period_in_epochs := paramsResponse.Params.ReductionPeriodInEpochs

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		response, err := stakingClient.Pool(
			context.Background(),
			&stakingtypes.QueryPoolRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get staking pool")
			return
		}

		bond_tokens := response.Pool.BondedTokens.Int64()

		apy := epoch_provision.Mul(staking).Mul(sdk.NewDecFromBigInt(big.NewInt(reduction_period_in_epochs))).Quo(sdk.NewDecFromBigInt(big.NewInt(bond_tokens)))

		if value, err := strconv.ParseFloat(apy.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse apy")
		} else {
			paramsAPYGauge.Set(value)
		}

		bankClient := banktypes.NewQueryClient(grpcConn)
		bankResponse, err := bankClient.TotalSupply(
			context.Background(),
			&banktypes.QueryTotalSupplyRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get bank total supply")
			return
		}

		for _, coin := range bankResponse.Supply {
			if supply, err := strconv.ParseInt(coin.Amount.String(), 10, 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not get total supply")
			} else {
				inflation := epoch_provision.Mul(sdk.NewDecFromBigInt(big.NewInt(reduction_period_in_epochs))).Quo(sdk.NewDecFromBigInt(big.NewInt(supply)))
				if inflationRate, err := strconv.ParseFloat(inflation.String(), 64); err != nil {
					sublogger.Error().
						Err(err).
						Msg("Could not parse apy")
				} else {
					paramsInflationGauge.Set(inflationRate)
				}
			}
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying global slashing params")
		queryStart := time.Now()

		slashingClient := slashingtypes.NewQueryClient(grpcConn)
		paramsResponse, err := slashingClient.Params(
			context.Background(),
			&slashingtypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get global slashing params")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying global slashing params")

		paramsDowntailJailDurationGauge.Set(paramsResponse.Params.DowntimeJailDuration.Seconds())
		paramsSignedBlocksWindowGauge.Set(float64(paramsResponse.Params.SignedBlocksWindow))

		if value, err := strconv.ParseFloat(paramsResponse.Params.MinSignedPerWindow.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse min signed per window")
		} else {
			paramsMinSignedPerWindowGauge.Set(value)
		}

		if value, err := strconv.ParseFloat(paramsResponse.Params.SlashFractionDoubleSign.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse slash fraction double sign")
		} else {
			paramsSlashFractionDoubleSign.Set(value)
		}

		if value, err := strconv.ParseFloat(paramsResponse.Params.SlashFractionDowntime.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse slash fraction downtime")
		} else {
			paramsSlashFractionDowntime.Set(value)
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying global distribution params")
		queryStart := time.Now()

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		paramsResponse, err := distributionClient.Params(
			context.Background(),
			&distributiontypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get global distribution params")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying global distribution params")

		// because cosmos's dec doesn't have .toFloat64() method or whatever and returns everything as int
		if value, err := strconv.ParseFloat(paramsResponse.Params.BaseProposerReward.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse base proposer reward")
		} else {
			paramsBaseProposerRewardGauge.Set(value)
		}

		if value, err := strconv.ParseFloat(paramsResponse.Params.BonusProposerReward.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse bonus proposer reward")
		} else {
			paramsBonusProposerRewardGauge.Set(value)
		}

		if value, err := strconv.ParseFloat(paramsResponse.Params.CommunityTax.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse community rate")
		} else {
			paramsCommunityTaxGauge.Set(value)
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying global stakeibc params")

		stakeibcClient := stakeibctypes.NewQueryClient(grpcConn)

		hostZoneAllResponse, err := stakeibcClient.HostZoneAll(
			context.Background(),
			&stakeibctypes.QueryAllHostZoneRequest{},
		)

		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get global stakeibc params")
			return
		}

		for _, hostZone := range hostZoneAllResponse.HostZone {
			if value, err := strconv.ParseFloat(hostZone.RedemptionRate.String(), 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not parse redemption rate")
			} else {
				paramsRedemptionRateGauge.With(prometheus.Labels{
					"chain": hostZone.ChainId,
				}).Set(value)
			}

			// Get the staked amount for each asset
			if value, err := strconv.ParseFloat(hostZone.StakedBal.String(), 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not parse staked amount")
			} else {
				var stakedAmount float64
				if hostZone.ChainId == "evmos_9001-2" {
					stakedAmount = float64(value / math.Pow10(18))
				} else {
					stakedAmount = float64(value / math.Pow10(6))
				}

				paramsStakedAmountGauge.With(prometheus.Labels{
					"chain": hostZone.ChainId,
				}).Set(stakedAmount)
			}
		}

	}()

	wg.Add(1)

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/params").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
