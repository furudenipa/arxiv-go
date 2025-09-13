# PR タイトル（Conventional Commits 準拠）
<!-- 例: feat(client): add rate limit option -->

## 概要
- このPRで解決すること/提供する価値を1〜3行で。

## 変更内容（What）
- 主な変更点を箇条書きで簡潔に。

## 背景・目的（Why）
- なぜ必要か、関連する文脈を簡潔に。

## 関連 Issue
- Closes #
- Relates to #

## 動作確認方法
1. 依存整理: `go mod tidy`
2. フォーマット: `go fmt ./...`（差分がないこと）
3. 静的解析: `go vet ./...`
4. ビルド: `go build ./...`
5. テスト: `go test -race -cover ./...`
6. 例の実行（必要に応じて）:
   - `go run example/search_example/main.go`
   - `go run example/iter_example/main.go`

## スクリーンショット / ログ（任意）
<!-- 実行結果、出力サンプル、失敗時ログなど -->

## 互換性・影響範囲
- 破壊的変更: あり / なし（詳細）
- 影響範囲（API/型/挙動）:

## セキュリティ / パフォーマンス / 可観測性
- 外部APIリスク、レート制限配慮、計測やログ追加の要不要など。

## テスト
- [ ] ユニットテストを追加/更新
- テスト対象・観点:

## チェックリスト（マージ前）
- [ ] PR タイトルが Conventional Commits
- [ ] `go fmt` / `go vet` / `go test -race -cover` を通過
- [ ] ドキュメント更新（必要なら `README.md` / `AGENTS.md`）
- [ ] サンプルコード更新（必要に応じて）
- [ ] 破壊的変更は明示し、移行手順を記載

## リリースノート（任意）
```release-note
短い箇条書きで変更点（利用者視点）
```

