package nude

import "testing"

func TestIsUrlNude(t *testing.T) {
	type args struct {
		imagePath string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "test if url nude image", args: struct{ imagePath string }{imagePath: "deleted"}, want: true, wantErr: false },
		{name: "test if url nude image", args: struct{ imagePath string }{imagePath: "deleted"}, want: false, wantErr: false },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsUrlNude(tt.args.imagePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsUrlNude() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsUrlNude() got = %v, want %v", got, tt.want)
			}
		})
	}
}
