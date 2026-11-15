Вот **один цельный Markdown-файл** со всеми запросами, переменными, кодами ошибок и примерами. Можешь сохранить его как `docs/graphql.md` или `API.md`.

```markdown
# GraphQL API

Базовый URL: `POST /query`  
Заголовок `Authorization: Bearer <token>` нужен для защищённых операций.

---

## Регистрация

```graphql
mutation Register {
  register(login: "alice", password: "secret123") {
    token
  }
}
```

## Логин

```graphql
mutation Login {
  login(login: "alice", password: "secret123") {
    token
  }
}
```

## Список постов (первая страница)

```graphql
query Posts {
  posts(first: 10) {
    items {
      id
      title
      body
      authorId
      commentsEnabled
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
```

## Список постов (следующая страница)

```graphql
query PostsNext {
  posts(first: 10, after: "10") {
    items {
      id
      title
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
```

## Один пост по ID

```graphql
query Post {
  post(id: "1") {
    id
    title
    body
    authorId
    commentsEnabled
  }
}
```

## Создать пост

```graphql
mutation CreatePost {
  createPost(input: {
    title: "Мой первый пост"
    body: "Привет, мир!"
    commentsEnabled: true
  }) {
    id
    title
    body
    authorId
    commentsEnabled
  }
}
```

## Включить или выключить комментарии

```graphql
mutation SetCommentsEnabled {
  setPostCommentsEnabled(postId: "1", enabled: false) {
    id
    title
    commentsEnabled
  }
}
```

## Комментарии к посту (первая страница)

```graphql
query PostComments {
  post(id: "1") {
    comments(first: 10) {
      items {
        id
        postId
        parentId
        authorId
        text
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

## Комментарии к посту (следующая страница)

```graphql
query PostCommentsNext {
  post(id: "1") {
    comments(first: 10, after: "5") {
      items {
        id
        text
        parentId
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

## Создать корневой комментарий

```graphql
mutation CreateRootComment {
  createComment(input: {
    postId: "1"
    text: "Отличный пост!"
  }) {
    id
    postId
    parentId
    authorId
    text
  }
}
```

## Ответить на комментарий

```graphql
mutation CreateReply {
  createComment(input: {
    postId: "1"
    parentId: "2"
    text: "Согласен!"
  }) {
    id
    postId
    parentId
    authorId
    text
  }
}
```

## Подписка на новые комментарии

```graphql
subscription CommentAdded {
  commentAdded(postId: "1") {
    id
    postId
    parentId
    authorId
    text
  }
}
```

---

## Переменные

### Создание поста

запрос:

```graphql
mutation CreatePost($input: CreatePostInput!) {
  createPost(input: $input) {
    id
    title
  }
}
```

переменные:

```json
{
  "input": {
    "title": "Заголовок",
    "body": "Текст поста",
    "commentsEnabled": true
  }
}
```

### Создание комментария

запрос:

```graphql
mutation CreateComment($input: CreateCommentInput!) {
  createComment(input: $input) {
    id
    text
  }
}
```

переменные:

```json
{
  "input": {
    "postId": "1",
    "parentId": "2",
    "text": "Комментарий"
  }
}
```

---

## Коды ошибок

| Код | Описание |
|---|---|
| `VALIDATION` | Неверный ID, пустое поле, слишком длинный текст, комментарии отключены |
| `NOT_FOUND` | Пост или комментарий не найден |
| `FORBIDDEN` | Нет прав (не автор поста, не авторизован) |
| `INTERNAL` | Внутренняя ошибка сервера |

### Пример ошибки

```json
{
  "errors": [
    {
      "message": "validation failed",
      "extensions": {
        "code": "VALIDATION"
      }
    }
  ],
  "data": null
}
```
```