package rbc

// func TestPurchaseParser(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&purchase{},
// 		filepath.Join("testdata", "purchase.json"),
// 		"test-purchase-id",
// 		expectedTransactionDetails{
// 			Account:     "************0000",
// 			Amount:      "26.19",
// 			Date:        time.Date(2025, time.June, 1, 7, 1, 46, 0, time.FixedZone("UTC-6", -6*60*60)),
// 			Currency:    "CAD",
// 			Direction:   domain.Out,
// 			Description: "UBER EATS",
// 		},
// 	)
// }

// func TestPurchaseParserForwarded(t *testing.T) {
// 	assertTransaction(
// 		t,
// 		&purchase{},
// 		filepath.Join("testdata", "purchase_fwd.json"),
// 		"test-purchase-fwd-id",
// 		expectedTransactionDetails{
// 			Account:     "************0000",
// 			Amount:      "19.12",
// 			Date:        time.Date(2025, time.May, 25, 19, 10, 29, 0, time.UTC),
// 			Currency:    "CAD",
// 			Direction:   domain.Out,
// 			Description: "SOME STORE",
// 		},
// 	)
// }
