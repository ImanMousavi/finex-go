package cron

import (
	"log"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/shopspring/decimal"
	"github.com/zsmartex/finex/config"
	"github.com/zsmartex/finex/models"
)

type ReleaseCommissionJob struct {
}

func (j *ReleaseCommissionJob) Process() {
	s := gocron.NewScheduler()
	s.Every(1).At("00:00:00").Do(releaseReferrals)
	<-s.Start()
}

type GroupReferral struct {
	FriendTrade uint64
	MemberID    uint64
}

type GroupUserReferral struct {
	Friend uint64
	UID    string
}

func releaseReferrals() {
	var group_referrals []*GroupReferral

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	config.DataBase.
		Model(&models.Commission{}).
		Select("COUNT(DISTINCT friend_uid) as friend_trade", "member_id").
		Where("CAST(\"created_at\" AS DATE) = ?", yesterday).
		Group("member_id").
		Find(&group_referrals)

	log.Println(group_referrals)

	for _, group_referral := range group_referrals {
		var commissions []*models.Commission

		earned_usdt := decimal.Zero

		config.DataBase.Where("member_id = ? AND CAST(\"created_at\" AS DATE) = ?", group_referral.MemberID, yesterday).Find(&commissions)

		for _, commission := range commissions {
			var currency *models.Currency

			config.DataBase.First(&currency, "id = ?", commission.CurrencyID)
			earned_usdt = earned_usdt.Add(currency.Price.Mul(commission.EarnAmount))
		}

		var btc_currency *models.Currency
		config.DataBase.First(&btc_currency, "id = ?", "btc")

		earned_btc := earned_usdt.DivRound(btc_currency.Price, 8)

		release_commission := &models.ReleaseCommission{
			AccountType: "spot",
			MemberID:    group_referral.MemberID,
			EarnedBTC:   earned_btc,
			FriendTrade: group_referral.FriendTrade,
			Friend:      0,
		}

		config.DataBase.Create(&release_commission)
	}

	var group_user_referrals []*GroupUserReferral

	config.DataBase.
		Model(&models.Member{}).
		Select("COUNT(referral_uid) as friend", "referral_uid as uid").
		Where("CAST(\"created_at\" AS DATE) = ?", yesterday).
		Group("referral_uid").
		Find(&group_user_referrals)

	for _, group_user_referral := range group_user_referrals {
		var member *models.Member
		var release_referral *models.ReleaseCommission

		config.DataBase.Where("uid = ?", group_user_referral.UID).Find(&member)
		if result := config.DataBase.Where("member_id = ? AND CAST(\"created_at\" AS DATE) = ?", member.ID, yesterday).First(&release_referral); result.Error == nil {
			config.DataBase.Model(&release_referral).Update("friend", group_user_referral.Friend)
		} else {
			release_commission := &models.ReleaseCommission{
				AccountType: "spot",
				MemberID:    member.ID,
				EarnedBTC:   decimal.Zero,
				FriendTrade: 0,
				Friend:      group_user_referral.Friend,
			}

			config.DataBase.Create(&release_commission)
		}
	}
}
