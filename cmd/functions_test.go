package main

import (
	"reflect"
	"testing"
	"time"
)

func TestGetUTCTimeFormat(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2023-02-01")
	expected := "2023-02-01T00:00:00.000Z"
	actual := getUTCTimeFormat(date)
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestConvertTransactions(t *testing.T) {
	type args struct {
		data [][]string
		ulid string
	}
	tests := []struct {
		name string
		args args
		want []Transaction
	}{
		{
			name: "basic conversion",
			args: args{
				data: [][]string{
					{"id", "date", "amount", "id_account"},
					{"1", "2023-02-01", "123.45", "123456789"},
				},
				ulid: "1234567890",
			},
			want: []Transaction{
				{
					Id:     "1",
					Date:   "2023-02-01",
					Amount: 123.45,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTransactions(tt.args.data, tt.args.ulid)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSummary(t *testing.T) {
	type args struct {
		transactionList []Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Testing get Summary",
			args: args{
				transactionList: []Transaction{
					{
						Id:     "1",
						Date:   "2023-02-01",
						Amount: 10.0,
					}, {
						Id:     "2",
						Date:   "2023-02-01",
						Amount: 12.45,
					}, {
						Id:     "3",
						Date:   "2023-03-01",
						Amount: -1.80,
					},
				},
			},
			want: []string{"Total balance is: 20.65", "Number of transactions in February: 2", "Number of transactions in March: 1",
				"Average debit amount: -1.80", "Average credit amount: 11.22"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSummary(tt.args.transactionList)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
