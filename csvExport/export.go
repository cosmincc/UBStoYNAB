package csvExport

import (
	"UBStoYNAB/ubsApi"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func ExportNormalAccountToCSV(transactions []ubsApi.CashAccountTrxes, accountAlias string) {
	var data = [][]string{{"Date", "Payee", "Memo", "Outflow", "Inflow"}}
	var missedLines = 0
	for _, element := range transactions {
		var date, payee, memo, outflow, inflow string
		trxDate, _ := time.Parse("02.01.2006", element.ValueDate)
		date = trxDate.Format("01/02/2006")
		if strings.HasPrefix(element.TrxAmount, "-") {
			outflow = strings.Trim(element.TrxAmount, "-")
			inflow = ""
		} else {
			outflow = ""
			inflow = element.TrxAmount
		}

		if strings.HasPrefix(element.Description, "KARTE") || (len(element.TransactionTextList) > 0 && strings.Contains(element.TransactionTextList[0], "Bezug")) {
			if len(element.TransactionTextList) > 0 && strings.HasPrefix(element.TransactionTextList[0], "Bezug") {
				payee = "Bargeld bezug"
				if len(element.TransactionTextList) > 1 {
					memo = element.TransactionTextList[1]
				}
			} else if element.TransactionTextList[0] == "Zahlung Maestro" && len(element.TransactionTextList) > 1 {
				payee = element.TransactionTextList[1]
			} else if element.TransactionTextList[0] == "Zahlung V PAY" && len(element.TransactionTextList) > 1 {
				payee = element.TransactionTextList[1]
			} else {
				payee = strings.Join(element.TransactionTextList, ", ")
			}
		} else if element.Description == "Dauerauftrag" {
			if len(element.PaymentInformationList) > 0 && len(element.PaymentInformationList[0].DescriptionList) > 0 {
				if element.PaymentInformationList[0].DescriptionList[0] == "Sparen" || element.TrxAmount == "-150.00" {
					payee = "Transfer to: UBS Sparkonto"
				} else if element.PaymentInformationList[0].DescriptionList[0] == "Fond" || element.TrxAmount == "-80.00" {
					payee = "Transfer to: UBS Fond"
				} else {
					payee = element.PaymentInformationList[0].DescriptionList[0]
				}
			}
		} else if element.Description == "Saldo DL-Preisabschluss" {
			payee = "UBS"
			memo = "UBS Gebühren"
		} else if element.Description == "Zinsabschluss" {
			payee = "UBS"
			memo = "UBS Zinsabschluss"
		} else if strings.Contains(element.Description, "Sammelauftrag") {
			for _, subelement := range element.PaymentInformationList {
				if strings.HasPrefix(subelement.Amount, "-") {
					outflow = strings.Trim(subelement.Amount, "-")
					inflow = ""
				} else {
					outflow = ""
					inflow = subelement.Amount
				}
				memo = "Sammelauftrag, Währung: " + subelement.Currency
				payee = subelement.DescriptionList[0]
				data = append(data, []string{date, payee, memo, outflow, inflow})
			}
			continue
		} else {
			if len(element.TransactionTextList) > 0 {
				if len(element.PaymentInformationList) > 0 && len(element.PaymentInformationList[0].DescriptionList) > 0 {
					if strings.Contains(element.PaymentInformationList[0].DescriptionList[0], "TWINT") || strings.Contains(element.TransactionTextList[0], "TWINT") {
						if strings.HasPrefix(element.TrxAmount, "-") {
							payee = "Twint Zahlung"
						} else {
							payee = "Twint Einzahlung"
						}

						if len(element.PaymentInformationList[0].DescriptionList) > 1 {
							memo = element.PaymentInformationList[0].DescriptionList[1] + ", " + element.TransactionTextList[0]
						} else {
							memo = element.PaymentInformationList[0].DescriptionList[0] + ", " + element.TransactionTextList[0]
						}
					} else {
						payee = element.PaymentInformationList[0].DescriptionList[0]
						memo = element.TransactionTextList[0]
					}
				}
			} else if len(element.PaymentInformationList) > 0 && len(element.PaymentInformationList[0].DescriptionList) > 0 {
				if strings.HasSuffix(element.PaymentInformationList[0].DescriptionList[0], "8754") {
					payee = "Payment to: UBS MasterCard"

				} else if strings.HasSuffix(element.PaymentInformationList[0].DescriptionList[0], "1707") {
					payee = "Payment to: UBS Visa"

				} else {
					payee = element.PaymentInformationList[0].DescriptionList[0]
					if len(element.PaymentInformationList[0].DescriptionList) > 1 {
						memo = element.PaymentInformationList[0].DescriptionList[len(element.PaymentInformationList[0].DescriptionList)-1]
					}
				}

			} else {
				missedLines++
				fmt.Println(element)
				fmt.Println("-------------")
			}

		}
		data = append(data, []string{date, payee, memo, outflow, inflow})
	}
	writeToFile(data, accountAlias)
	fmt.Printf("%d von %d Transaktionen wurden exportiert\n", (len(transactions) - missedLines), len(transactions))
}

func ExportCreditCardToCSV(cardTransactions []ubsApi.CardTransactions, accountTransactions []ubsApi.CreditCardAccountTransactions, cardName string) {
	var data = [][]string{{"Date", "Payee", "Memo", "Outflow", "Inflow"}}
	for _, element := range cardTransactions {
		var date, payee, memo, outflow, inflow string
		trxDate, _ := time.Parse("20060102", element.TransactionDate)
		date = trxDate.Format("01/02/2006")
		if strings.HasPrefix(element.PostingAmount, "-") {
			outflow = strings.Trim(element.PostingAmount, "-")
			inflow = ""
		} else {
			outflow = ""
			inflow = element.PostingAmount
		}

		if strings.HasPrefix(element.TransactionText, "PAYPAL") {
			payee = strings.Split(strings.Replace(element.TransactionText, "PAYPAL *", "", -1), "  ")[0]
			memo = "Paypal"
		} else if strings.Contains(element.TransactionText, "ZUSCHLAG") || strings.Contains(element.TransactionText, "STORNO ZUSCHL") {
			payee = "UBS Gebühren"
			memo = element.TransactionText
		} else if strings.Contains(element.TransactionText, "  ") {
			payee = strings.Split(element.TransactionText, "  ")[0]
		} else {
			payee = element.TransactionText
		}
		data = append(data, []string{date, payee, memo, outflow, inflow})
	}

	for _, element := range accountTransactions {
		var date, payee, memo, outflow, inflow string
		trxDate, _ := time.Parse("20060102", element.TransactionDate)
		date = trxDate.Format("01/02/2006")
		if strings.HasPrefix(element.PostingAmount, "-") {
			outflow = strings.Trim(element.PostingAmount, "-")
			inflow = ""
		} else {
			outflow = ""
			inflow = element.PostingAmount
		}

		if strings.HasPrefix(element.TransactionText, "BANKVERGUETUNG") {
			payee = "Payment from: UBS Privatkonto"
		} else {
			payee = element.TransactionText
		}
		data = append(data, []string{date, payee, memo, outflow, inflow})
	}

	writeToFile(data, cardName)
	fmt.Printf("%d Transaktionen wurden exportiert\n", (len(accountTransactions) + len(cardTransactions)))

}

func writeToFile(data [][]string, fileName string) {
	file, err := os.Create(fileName + ".csv")
	checkError("Kann Datei nicht erstellen", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		checkError("Kann nicht in Datei schreiben", err)
	}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
