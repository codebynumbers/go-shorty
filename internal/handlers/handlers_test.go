package handlers

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	"github.com/codebynumbers/go-shorty/internal/connections"
	"github.com/elliotchance/redismock"
	"github.com/go-redis/redis"
	"testing"
)

func TestGenerateHash(t *testing.T) {
	// test stuff here...
	cases := []struct {
		in, want string
	}{
		{"", "811c9dc5"},
		{"/", "2a0c975e"},
		{"http://www.stuff.com", "654d9cc5"},
	}
	for _, c := range cases {
		got := generateHash(c.in)
		if got != c.want {
			t.Errorf("generateHash(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func newTestRedis() *redismock.ClientMock {

	client := redis.NewClient(&redis.Options{
		Addr: "99.99.99.99:666",
	})

	return redismock.NewNiceMock(client)
}

func TestCachedGetUrlFoundInRedis(t *testing.T) {

	in := "abcd"
	want := "http://abcd.com"

	config := configuration.Configure()
	db := connections.InitDb(config)
	mockCache := newTestRedis()
	mockCache.On("Get").Return(redis.NewStringResult(want, nil))

	env := HandlerEnv{
		AppConfig: config,
		Db:        db,
		Cache:     mockCache,
	}

	got, _ := env.cachedGetUrl(in)
	if got != want {
		t.Errorf("cachedGetUrl(%q) == %q, want %q", in, got, want)
	}

}

func TestCachedGetUrlNotFoundInRedis(t *testing.T) {

	in := "abcd"
	want := "http://abcd.com"

	config := configuration.Configure()
	mockCache := newTestRedis()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"url"}).AddRow(want)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	env := HandlerEnv{
		AppConfig: config,
		Db:        db,
		Cache:     mockCache,
	}

	got, _ := env.cachedGetUrl(in)
	if got != want {
		t.Errorf("cachedGetUrl(%q) == %q, want %q", in, got, want)
	}
}
