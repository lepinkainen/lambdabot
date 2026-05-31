package command

import (
	"regexp"
	"testing"
)

func TestTVMaze(t *testing.T) {
	type args struct {
		args string
	}
	tests := []struct {
		name    string
		args    args
		want    *regexp.Regexp
		wantErr bool
	}{
		{"Obi-Wan Kenobi", args{args: "obi wan kenobi"}, regexp.MustCompile(`^Latest episode of Obi-Wan Kenobi 1x06 'Part VI' airs 2022-06-22 \([^)]+\) on Disney\+ \[Ended\]$`), false},
		{"Gilmore Girls", args{args: "gilmore girls"}, regexp.MustCompile(`^Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 \([^)]+\) on The CW \[Ended\]$`), false},
		{"The Grand Tour", args{args: "grand tour"}, regexp.MustCompile(`^Latest episode of The Grand Tour 6x01 'The Grand Tour: One for the Road' airs 2024-09-13 \([^)]+\) on Prime Video( \[Ended\])?$`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := TVMaze(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("TVMaze() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.MatchString(got) {
				t.Errorf("TVMaze() = '%v', want match '%v'", got, tt.want)
			}
		})
	}
}
