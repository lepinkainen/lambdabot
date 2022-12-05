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
		{"Obi-Wan Kenobi", args{args: "obi wan kenobi"}, "Latest episode of Obi-Wan Kenobi 1x06 'Part VI' airs 2022-06-22 (5 months ago) on Disney+ [Ended]", false},
		{"Gilmore Girls", args{args: "gilmore girls"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 (15 years ago) on The CW [Ended]", false},
		{"The Grand Tour", args{args: "grand tour"}, "Latest episode of The Grand Tour 5x01 'The Grand Tour Presents: A Scandi Flick' airs 2022-09-16 (2 months ago) on Prime Video", false},
		{"Doctor Who", args{args: "doctor who"}, "Latest episode of Doctor Who 13x06 'Chapter Six: The Vanquishers' airs 2021-12-05 (1 year ago) on BBC One", false},
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
				t.Errorf("TVMaze() = %v, want %v", got, tt.want)
			}
		})
	}
}
