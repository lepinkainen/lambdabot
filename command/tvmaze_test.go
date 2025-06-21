package command

import "testing"

func TestTVMaze(t *testing.T) {
	type args struct {
		args string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Obi-Wan Kenobi", args{args: "obi wan kenobi"}, "Latest episode of Obi-Wan Kenobi 1x06 'Part VI' airs 2022-06-22 (3 years ago) on Disney+ [Ended]", false},
		{"Gilmore Girls", args{args: "gilmore girls"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 (18 years ago) on The CW [Ended]", false},
		{"The Grand Tour", args{args: "grand tour"}, "Latest episode of The Grand Tour 6x01 'The Grand Tour: One for the Road' airs 2024-09-13 (9 months ago) on Prime Video [Ended]", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := TVMaze(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("TVMaze() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TVMaze() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
