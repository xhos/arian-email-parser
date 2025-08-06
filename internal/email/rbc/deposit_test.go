package rbc

// func TestDepositParser(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&deposit{},
// 		filepath.Join("testdata", "deposit.json"),
// 		"test-deposit-id",
// 		expectedTransactionDetails{
// 			Account:     "Savings",
// 			Amount:      "42000.16",
// 			Date:        time.Date(2025, time.May, 31, 10, 27, 46, 0, time.FixedZone("UTC-6", -6*60*60)),
// 			Currency:    "CAD",
// 			Direction:   domain.In,
// 			Description: "RBC Deposit",
// 		},
// 	)
// }

// func TestDepositParserForwarded(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&deposit{},
// 		filepath.Join("testdata", "deposit_fwd.json"),
// 		"test-deposit-fwd-id",
// 		expectedTransactionDetails{
// 			Account:     "Daily",
// 			Amount:      "60.00",
// 			Date:        time.Date(2025, time.May, 27, 0, 0, 0, 0, time.UTC),
// 			Currency:    "CAD",
// 			Direction:   domain.In,
// 			Description: "RBC Deposit",
// 		},
// 	)
// }
