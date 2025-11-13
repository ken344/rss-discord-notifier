package models

import "time"

// State は、アプリケーションの状態を表すモデル
// state.json ファイルに保存される
type State struct {
	// Version は状態ファイルのバージョン
	Version string `json:"version"`

	// LastUpdate は最後に状態を更新した日時
	LastUpdate time.Time `json:"last_update"`

	// Feeds は各フィードごとの状態
	Feeds map[string]*FeedState `json:"feeds"`

	// Statistics は統計情報
	Statistics *Statistics `json:"statistics"`
}

// FeedState は、個別のフィードの状態を表すモデル
type FeedState struct {
	// LastCheck は最後にこのフィードをチェックした日時
	LastCheck time.Time `json:"last_check"`

	// NotifiedArticles は通知済みの記事のリスト
	NotifiedArticles []*NotifiedArticle `json:"notified_articles"`
}

// NotifiedArticle は、通知済みの記事を表すモデル
type NotifiedArticle struct {
	// ID は記事の一意な識別子
	ID string `json:"id"`

	// Title は記事のタイトル
	Title string `json:"title"`

	// URL は記事のURL
	URL string `json:"url"`

	// PublishedAt は記事の公開日時
	PublishedAt time.Time `json:"published_at"`

	// NotifiedAt は通知を送信した日時
	NotifiedAt time.Time `json:"notified_at"`
}

// Statistics は、アプリケーションの統計情報を表すモデル
type Statistics struct {
	// TotalArticlesNotified は通知した記事の総数
	TotalArticlesNotified int `json:"total_articles_notified"`

	// TotalFeedsChecked はチェックしたフィードの総数
	TotalFeedsChecked int `json:"total_feeds_checked"`

	// LastRunDuration は最後の実行にかかった時間（秒）
	LastRunDuration float64 `json:"last_run_duration"`
}

// NewState は、新しい状態を作成する
func NewState() *State {
	return &State{
		Version:    "1.0",
		LastUpdate: time.Now(),
		Feeds:      make(map[string]*FeedState),
		Statistics: &Statistics{
			TotalArticlesNotified: 0,
			TotalFeedsChecked:     0,
			LastRunDuration:       0,
		},
	}
}

// GetFeedState は、指定されたフィードURLの状態を取得する
// 存在しない場合は新しく作成する
func (s *State) GetFeedState(feedURL string) *FeedState {
	if s.Feeds == nil {
		s.Feeds = make(map[string]*FeedState)
	}

	if feedState, exists := s.Feeds[feedURL]; exists {
		return feedState
	}

	// 存在しない場合は新しく作成
	feedState := &FeedState{
		LastCheck:        time.Now(),
		NotifiedArticles: make([]*NotifiedArticle, 0),
	}
	s.Feeds[feedURL] = feedState
	return feedState
}

// IsArticleNotified は、指定された記事IDが通知済みかチェックする
func (fs *FeedState) IsArticleNotified(articleID string) bool {
	for _, article := range fs.NotifiedArticles {
		if article.ID == articleID {
			return true
		}
	}
	return false
}

// AddNotifiedArticle は、通知済み記事を追加する
func (fs *FeedState) AddNotifiedArticle(article *Article) {
	notifiedArticle := &NotifiedArticle{
		ID:          article.ID,
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: article.PublishedAt,
		NotifiedAt:  time.Now(),
	}

	fs.NotifiedArticles = append(fs.NotifiedArticles, notifiedArticle)
	fs.LastCheck = time.Now()
}

// CleanupOldArticles は、古い記事情報を削除する（メモリ節約）
// daysOld より古い記事を削除
func (fs *FeedState) CleanupOldArticles(daysOld int) {
	if daysOld <= 0 {
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	newList := make([]*NotifiedArticle, 0)

	for _, article := range fs.NotifiedArticles {
		// cutoffDate より新しい記事のみ保持
		if article.NotifiedAt.After(cutoffDate) {
			newList = append(newList, article)
		}
	}

	fs.NotifiedArticles = newList
}

// LimitArticleCount は、通知済み記事の数を制限する
// 最新のN件のみを保持
func (fs *FeedState) LimitArticleCount(maxCount int) {
	if maxCount <= 0 || len(fs.NotifiedArticles) <= maxCount {
		return
	}

	// 新しい順にソート済みと仮定して、最新のmaxCount件のみを保持
	// （実際の実装では、NotifiedAtでソートしてから最新N件を取るのが望ましい）
	fs.NotifiedArticles = fs.NotifiedArticles[len(fs.NotifiedArticles)-maxCount:]
}
