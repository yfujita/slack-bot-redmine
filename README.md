# slack-bot-redmine
自分のissueをslackに流す

使い方：
```
go run main.go -webHookUrl https://hooks.slack.com/services/xxxxxx -channel "#チャンネル名" -botName bot名 -botIcon :アイコン名: -redmineHost https://ホスト -redmineKey RedmineのAPIキー -redmineUserId ユーザーID
```
