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
		{"Mandalorean", args{args: "mandalorean"}, "Next episode of The Mandalorian 2x2 'Chapter 10: The Confrontation' airs 2020-11-06 on Disney+", false},
		{"gilmore girls", args{args: "gilmore girls"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 on The CW [Ended]", false},
		{"grand tour", args{args: "grand tour"}, "Latest episode of Gilmore Girls 7x22 'Bon Voyage' airs 2007-05-15 on The CW [Ended]", false},
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
