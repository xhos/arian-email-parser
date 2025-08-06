package rbc

// func TestPaymentParser(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&payment{},
// 		filepath.Join("testdata", "payment.json"),
// 		"test-payment-id",
// 		expectedTransactionDetails{
// 			Account:     "************0000",
// 			Amount:      "1500.00",
// 			Date:        time.Date(2025, time.June, 10, 17, 43, 28, 0, time.FixedZone("UTC-6", -6*60*60)),
// 			Currency:    "CAD",
// 			Direction:   domain.In,
// 			Description: "RBC Payment",
// 		},
// 	)
// }

// func TestPaymentParserForwarded(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&payment{},
// 		filepath.Join("testdata", "payment_fwd.json"),
// 		"test-payment-fwd-id",
// 		expectedTransactionDetails{
// 			Account:     "************0000",
// 			Amount:      "1500.00",
// 			Date:        time.Date(2025, time.June, 10, 0, 0, 0, 0, time.UTC),
// 			Currency:    "CAD",
// 			Direction:   domain.In,
// 			Description: "RBC Payment",
// 		},
// 	)
// }
