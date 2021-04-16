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
		//{"Mandalorean", args{args: "mandalorean"}, "Next episode of The Mandalorian 2x02 'Chapter 10: The Confrontation' airs 2020-11-06 (5 days from now) on Disney+", false},
		{"Gilmore Girls", args{args: "gilmore girls"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 (14 years ago) on The CW [Ended]", false},
		{"The Grand Tour", args{args: "grand tour"}, "Latest episode of The Grand Tour 4x02 'The Grand Tour Presents: A Massive Hunt' airs 2020-12-17 (3 months ago) on Amazon Prime Video", false},
		{"Doctor Who", args{args: "doctor who"}, "Latest episode of Doctor Who 12x10 'The Timeless Children' airs 2020-03-01 (1 year ago) on BBC One", false},
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
