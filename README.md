# RL Arena

## Game Server
### Database
Create migration:
```
goose -dir=migrations create rlarena sql
```
Create Database from migrations
```
goose -dir=migrations sqlite3 app.db up
```