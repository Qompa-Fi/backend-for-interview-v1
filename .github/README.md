
# Backend for Technical interview V1

## Endpoints

### `GET /ws/tasks` - Open a WebSocket connection to read the tasks for the current workspace

#### Query parameters

- **api_key** - API key that identifies the caller

   *Type:* string

   *Required*: true

- **workspace_id**- ID of the current workspace

   *Type:* string

   *Required*: false

   *Format:* `^[A-z\-\_\d]+$`

   *Default*: `default`

#### Response body

Streamed lines of **Array&lt;Task&gt;**.

##### Example

```json
[]
[{"id":2,"name":"abc","type":"gx.tiny","status":"running"}]
[{"id":2,"name":"abc","type":"gx.tiny","status":"completed"}]
[{"id":2,"name":"abc","type":"gx.tiny","status":"completed"}]
```

#### cURL example

```shell
# asdas
$ websocat 'wss://<host>/ws/tasks?api_key=<api_key>&workspace_id=<workspace_id>'
```

### `GET /tasks` - Get the list of tasks for the current workspace

#### Query parameters

- **api_key** - API key that identifies the caller

   *Type:* string

   *Required*: true

- **workspace_id**- ID of the current workspace

   *Type:* string

   *Required*: false

   *Format:* `^[A-z\-\_\d]+$`

   *Default*: `default`

#### Response body

**{
    "tasks": Array&lt;Task&gt;
}**.

##### Example

```json
{"tasks":[{"id":3,"name":"abc","type":"gx.tiny","status":"running"}]}
```

### `POST /tasks` - Create a new task

#### Query parameters

- **api_key** - API key that identifies the caller

   *Type:* string

   *Required*: true

- **workspace_id**- ID of the current workspace

   *Type:* string

   *Required*: false

   *Format:* `^[A-z\-\_\d]+$`

   *Default*: `default`

#### Request body

- **name** - Name of the task

  *Type:* string

  *Required*: true

- **type** - Type of the task

  *Type:* string

  *Required*: true

  *Format:* `^[A-z\-\_\d]+$`

  *Default*: `default`

  *Available values:* `gx.tiny`, `gx.micro`, `gx.small`, `gx.medium`, `gx.large`, `gx.heavy`

#### Response body

Item of type **{
    "task": &lt;Task&gt;
}**.

##### Example

```json
{"task":{"id":3,"name":"abc","type":"gx.tiny","status":"queued"}}
```

#### cURL example

```shell
# asdas
$ curl -X POST 'https://<host>/tasks?api_key=<api_key>&workspace_id=<workspace_id>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Task name",
    "type": "gx.tiny"
  }'
```

### `DELETE /tasks/:id` - Delete a queued task in the current workspace

#### Query parameters

- **api_key** - API key that identifies the caller

   *Type:* string

   *Required*: true

- **workspace_id**- ID of the current workspace

   *Type:* string

   *Required*: false

   *Format:* `^[A-z\-\_\d]+$`

   *Default*: `default`

#### Response body

Nothing but *204 No Content*.

#### cURL example

```shell
# asdas
$ curl -X DELETE 'https://<host>/tasks/<id>?api_key=<api_key>&workspace_id=<workspace_id>'
```

### `POST /tasks/flush` - Flush all dispatched tasks in the current workspace

#### Query parameters

- **api_key** - API key that identifies the caller

   *Type:* string

   *Required*: true

- **workspace_id**- ID of the current workspace

   *Type:* string

   *Required*: false

   *Format:* `^[A-z\-\_\d]+$`

   *Default*: `default`

#### Response body

Nothing but *204 No Content*.

#### cURL example

```shell
# asdas
$ curl -X POST 'https://<host>/tasks/flush?api_key=<api_key>&workspace_id=<workspace_id>'
```
