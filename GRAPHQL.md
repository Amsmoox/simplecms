# GraphQL in AcaDelta CMS

This project uses GraphQL to provide a flexible API for your CMS data. The implementation uses [gqlgen](https://github.com/99designs/gqlgen), a Go library that generates code from a GraphQL schema.

## Getting Started

### Using the GraphQL Playground

1. Start your Buffalo application: 
   ```
   buffalo dev
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:3000/playground
   ```

3. You'll see the GraphQL Playground interface where you can:
   - Write and execute queries/mutations
   - Browse schema documentation
   - View query history

## Available Operations

The current schema defines:

- **Queries**:
  - `todos`: Returns a list of all todo items

- **Mutations**:
  - `createTodo`: Creates a new todo item

### Example Query

```graphql
query {
  todos {
    id
    text
    done
    user {
      id
      name
    }
  }
}
```

### Example Mutation

```graphql
mutation {
  createTodo(input: {
    text: "Learn GraphQL"
    userId: "1"
  }) {
    id
    text
    done
  }
}
```

## Customizing the Schema

To modify the GraphQL schema:

1. Edit `graph/schema.graphqls`
2. Run code generation:
   ```
   go run github.com/99designs/gqlgen generate
   ```
   (or if you have PATH issues: `/path/to/go/bin/gqlgen generate`)

3. Implement the newly generated resolver methods in `graph/schema.resolvers.go`

## Extending the Schema for CMS

For a CMS, you'll likely want to add types for your content:

```graphql
type Page {
  id: ID!
  title: String!
  slug: String!
  content: String!
  published: Boolean!
  createdAt: String!
  updatedAt: String!
}

type Query {
  # Existing queries...
  pages: [Page!]!
  page(id: ID!): Page
  pageBySlug(slug: String!): Page
}

input NewPage {
  title: String!
  slug: String! 
  content: String!
  published: Boolean!
}

input UpdatePage {
  title: String
  slug: String
  content: String
  published: Boolean
}

type Mutation {
  # Existing mutations...
  createPage(input: NewPage!): Page!
  updatePage(id: ID!, input: UpdatePage!): Page!
  deletePage(id: ID!): Boolean!
}
```

## Integration with Buffalo

The GraphQL integration works as follows:

- GraphQL endpoint: `POST /query`
- GraphQL Playground UI: `GET /playground`

The handlers are defined in `actions/graphql.go` and registered in `actions/app.go`.

**Note:** CSRF protection is disabled for these endpoints to allow the GraphQL Playground to work correctly. In a production environment, you should consider implementing an alternative security approach such as:
- Token-based authentication
- GraphQL-specific authentication middleware
- Origin/Referer checking

## Database Integration

To use your Buffalo database models with GraphQL:

1. Define GraphQL types in `graph/schema.graphqls`
2. In your resolvers (`graph/schema.resolvers.go`), use Buffalo's context to access the transaction:
   ```go
   func (r *queryResolver) Pages(ctx context.Context) ([]*model.Page, error) {
       // Get the Buffalo transaction from context
       c := ctx.Value("buffalo_ctx").(buffalo.Context)
       tx := c.Value("tx").(*pop.Connection)
       
       // Use the models package to query data
       pages := []models.Page{}
       err := tx.All(&pages)
       if err != nil {
           return nil, err
       }
       
       // Map from database models to GraphQL models
       result := make([]*model.Page, len(pages))
       for i, p := range pages {
           result[i] = &model.Page{
               ID:        p.ID.String(),
               Title:     p.Title,
               Slug:      p.Slug,
               Content:   p.Content,
               Published: p.Published,
               CreatedAt: p.CreatedAt.String(),
               UpdatedAt: p.UpdatedAt.String(),
           }
       }
       
       return result, nil
   }
   ```

## Authentication and Authorization

For protected GraphQL operations, you can use Buffalo's authentication middleware and access the current user in your resolvers:

```go
func (r *mutationResolver) CreatePage(ctx context.Context, input model.NewPage) (*model.Page, error) {
    c := ctx.Value("buffalo_ctx").(buffalo.Context)
    
    // Check if user is authenticated
    currentUser := c.Value("current_user")
    if currentUser == nil {
        return nil, fmt.Errorf("unauthorized")
    }
    
    // Proceed with page creation...
}
``` 