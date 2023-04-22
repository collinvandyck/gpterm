package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/collinvandyck/gpterm/db"
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/errs"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
)

const (
	DBName                   = "gpterm.db"
	ConfigChatMessageContext = "chat.message-context"
	CredentialAPIKey         = "api_key"
	CredentialGithubToken    = "github_token"
)

func DefaultStorePath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	res := filepath.Join(hd, ".config", "gpterm")
	return res, nil
}

func DefaultDBPath() (string, error) {
	sp, err := DefaultStorePath()
	if err != nil {
		return "", err
	}
	res := filepath.Join(sp, DBName)
	return res, nil
}

type Store struct {
	log.Logger
	dir     string
	db      *sql.DB
	queries *query.Queries
}

type StoreOpt func(*Store)

func StoreLog(log log.Logger) StoreOpt {
	return func(s *Store) {
		s.Logger = log
	}
}

func StoreDir(path string) StoreOpt {
	return func(s *Store) {
		s.dir = path
	}
}

func New(opts ...StoreOpt) (*Store, error) {
	store := &Store{
		Logger: log.Discard,
	}
	for _, o := range opts {
		o(store)
	}
	if store.dir == "" {
		dir, err := DefaultStorePath()
		if err != nil {
			return nil, err
		}
		store.dir = dir
	}
	err := store.init()
	if err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	return store, nil
}

var ErrNoMoreConversations = errors.New("no more conversations")

func (s *Store) GetConfig(ctx context.Context) (Config, error) {
	return s.queries.GetConfig(ctx)
}

func (s *Store) SetConfigInt(ctx context.Context, name string, value int) error {
	if value < 0 {
		return errors.New("value must be greater than 0")
	}
	return s.queries.SetConfigValue(ctx, query.SetConfigValueParams{
		Name:  name,
		Value: strconv.Itoa(value),
	})
}

func (s *Store) GetConfigInt(ctx context.Context, name string, defaultValue int) (int, error) {
	val, err := s.queries.GetConfigValue(ctx, name)
	switch {
	case errs.IsDBNotFound(err):
		return defaultValue, nil
	case err != nil:
		return -1, err
	default:
		return strconv.Atoi(val)
	}
}

func (s *Store) NextConversation(ctx context.Context) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queryTX := s.queries.WithTx(tx)

	c, err := queryTX.NextConversation(ctx)
	switch {
	case err == nil:
		err = queryTX.UnsetSelectedConversation(ctx)
		if err != nil {
			return err
		}
		err = queryTX.SetSelectedConversation(ctx, c.ID)
		if err != nil {
			return err
		}
		return tx.Commit()
	case errs.IsDBNotFound(err):
		c, err = queryTX.GetActiveConversation(ctx)
		if err != nil {
			return err
		}
		count, err := queryTX.CountMessagesForConversation(ctx, c.ID)
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrNoMoreConversations
		}
		c, err = queryTX.CreateConversation(ctx)
		if err != nil {
			return err
		}
		err = queryTX.UnsetSelectedConversation(ctx)
		if err != nil {
			return err
		}
		err = queryTX.SetSelectedConversation(ctx, c.ID)
		if err != nil {
			return err
		}
		return tx.Commit()
	default:
		return err
	}
}

func (s *Store) PreviousConversation(ctx context.Context) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	queryTX := s.queries.WithTx(tx)

	c, err := queryTX.PreviousConversation(ctx)
	switch {
	case err == nil:
		err = queryTX.UnsetSelectedConversation(ctx)
		if err != nil {
			return err
		}
		err = queryTX.SetSelectedConversation(ctx, c.ID)
		if err != nil {
			return err
		}
		return tx.Commit()
	case errs.IsDBNotFound(err):
		return ErrNoMoreConversations
	default:
		return err
	}
}

func (s *Store) GetTotalUsage(ctx context.Context) (res query.Usage, err error) {
	completion, err := s.queries.GetCompletionTokens(ctx)
	if err != nil {
		return
	}
	prompt, err := s.queries.GetPromptTokens(ctx)
	if err != nil {
		return
	}
	total, err := s.queries.GetTotalTokens(ctx)
	if err != nil {
		return
	}
	res = query.Usage{
		PromptTokens:     int64(prompt),
		CompletionTokens: int64(completion),
		TotalTokens:      int64(total),
	}
	return
}

func (s *Store) GetPreviousMessageForRole(ctx context.Context, role string, offset int) (query.Message, error) {
	if offset <= 0 {
		return query.Message{}, errors.New("bad offset")
	}
	return s.queries.GetPreviousMessageForRole(ctx, query.GetPreviousMessageForRoleParams{
		Role:   role,
		Offset: int64(offset - 1),
	})
}

func (s *Store) GetLastMessages(ctx context.Context, count int) ([]query.Message, error) {
	return s.queries.GetLatestMessages(ctx, int64(count))
}

func (s *Store) SaveRequest(ctx context.Context, req openai.ChatCompletionRequest) error {
	if len(req.Messages) > 1 {
		m := req.Messages[len(req.Messages)-1]
		err := s.queries.InsertMessage(ctx, query.InsertMessageParams{
			Role:    m.Role,
			Content: strings.TrimSpace(m.Content),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveStreamResults(ctx context.Context, text string, usage openai.Usage, failure error) error {
	err := s.queries.InsertMessage(ctx, query.InsertMessageParams{
		Role:    "assistant",
		Content: strings.TrimSpace(text),
	})
	if err != nil {
		return err
	}
	// only record usage if it was reported.
	if usage.PromptTokens+usage.CompletionTokens+usage.TotalTokens > 0 {
		err = s.queries.InsertUsage(ctx, query.InsertUsageParams{
			PromptTokens:     int64(usage.PromptTokens),
			CompletionTokens: int64(usage.CompletionTokens),
			TotalTokens:      int64(usage.TotalTokens),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveRequestResponse(ctx context.Context, req openai.ChatCompletionRequest, resp openai.ChatCompletionResponse) error {
	// save the last message in the request. that skips the context messages
	err := s.SaveRequest(ctx, req)
	if err != nil {
		return err
	}
	// save all responses
	for _, choice := range resp.Choices {
		m := choice.Message
		err := s.queries.InsertMessage(ctx, query.InsertMessageParams{
			Role:    m.Role,
			Content: strings.TrimSpace(m.Content),
		})
		if err != nil {
			return err
		}
	}
	// save usage
	err = s.queries.InsertUsage(ctx, query.InsertUsageParams{
		PromptTokens:     int64(resp.Usage.PromptTokens),
		CompletionTokens: int64(resp.Usage.CompletionTokens),
		TotalTokens:      int64(resp.Usage.TotalTokens),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetCredential(ctx context.Context, name string) (string, error) {
	res, err := s.queries.GetCredential(ctx, name)
	switch {
	case errs.IsDBNotFound(err):
	case err != nil:
		return "", err
	}
	return res, nil
}

func (s *Store) SetCredential(ctx context.Context, name string, value string) error {
	return s.queries.UpdateCredential(ctx, query.UpdateCredentialParams{
		Name:  name,
		Value: value,
	})
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init() error {
	if err := ensureDir(s.dir); err != nil {
		return err
	}
	if err := s.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := s.initDB(); err != nil {
		return fmt.Errorf("initDB: %w", err)
	}
	s.queries = query.New(s.db)
	return nil
}

func (s *Store) migrate() error {
	sourceDriver, err := iofs.New(db.FSMigrations, "migrations")
	if err != nil {
		return err
	}
	path := "sqlite3://" + s.DBPath()
	mg, err := migrate.NewWithSourceInstance("iofs", sourceDriver, path)
	if err != nil {
		return err
	}
	err = mg.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
	case err != nil:
		return fmt.Errorf("up: %w", err)
	}
	return nil
}

func (s *Store) initDB() error {
	db, err := sql.Open("sqlite3", s.DBPath())
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Store) DBPath() string {
	return filepath.Join(s.dir, DBName)
}

func ensureDir(dir string) error {
	info, err := os.Stat(dir)
	switch {
	case os.IsNotExist(err):
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case info.IsDir():
	case !info.IsDir():
		return fmt.Errorf("%q exists but is a file", dir)
	}
	return nil
}
