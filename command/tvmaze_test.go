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
		{"Gilmore Girls", args{args: "gilmore girls"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 (13 years ago) on The CW [Ended]", false},
		{"The Grand Tour", args{args: "grand tour"}, "Latest episode of The Grand Tour 4x02 'The Grand Tour Presents: Madagascar Special' airs [UNKNOWN] on Amazon Prime", false},
		{"Doctor Who", args{args: "doctor who"}, "Next episode of Doctor Who 12x00 'Revolution of the Daleks' airs 2020-12-25 (1 month from now) on BBC One", false},
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
