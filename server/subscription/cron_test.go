package subscription

import "testing"

func TestValidateCronDefault(t *testing.T) {
	if err := ValidateCron(DefaultCron); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCronRejectsInvalid(t *testing.T) {
	if err := ValidateCron("not a cron"); err == nil {
		t.Fatal("expected invalid cron to fail")
	}
}

func TestSourceCronDefault(t *testing.T) {
	if got := SourceCron(""); got != DefaultCron {
		t.Fatalf("SourceCron empty = %q, want %q", got, DefaultCron)
	}
	if got := SourceCron("  */30 * * * *  "); got != "*/30 * * * *" {
		t.Fatalf("SourceCron trimmed = %q", got)
	}
}

func TestFormatNowUsesShanghai(t *testing.T) {
	got := FormatNow()
	if len(got) != 19 {
		t.Fatalf("unexpected time format %q", got)
	}
}
