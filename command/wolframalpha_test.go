//go:build !ci

package command

import "testing"

func TestWolframAlpha(t *testing.T) {
	type args struct {
		args string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"1 kg in pounds", args{args: "1kg in pounds"}, "1kg in pounds = 2.205 lb  (pounds)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WolframAlpha(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("WolframAlpha() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WolframAlpha() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
