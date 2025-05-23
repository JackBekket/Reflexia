# pkg/store

## Overview
The `pkg/store` package provides functionality for initializing and managing a vector store using PostgreSQL with the `pgvector` extension, integrated with OpenAI's embedding capabilities. It focuses on creating a secure, structured storage system for vector data with built-in protection against accidental data loss.

## Features
- **Vector Store Initialization**: Creates a PostgreSQL-backed vector store using `pgvector`.
- **OpenAI Integration**: Uses OpenAI's `text-embedding-ada-002` model for embedding operations.
- **Pre-Delete Protection**: Ensures accidental deletion of entire collections is prevented by default.
- **Context Awareness**: Supports context cancellation and timeout control across all operations.

## External Configurations
The package behavior is controlled by the following parameters:

| Configuration         | Description                              | Required? | Notes                              |
|----------------------|------------------------------------------|-----------|------------------------------------|
| `dbURL`              | PostgreSQL connection string               | Yes       | Must be a valid PostgreSQL URL     |
| `aiURL`              | OpenAI API endpoint URL                    | Yes       | e.g., `https://api.openai.com/v1`  |
| `aiToken`            | OpenAI API authentication token            | Yes       | Requires valid OpenAI API key      |
| `name`               | Vector store collection name             | Yes       | Identifies the collection in DB  |
| **Fixed Settings**   |                                          |           |                                    |
| Embedding Model      | `text-embedding-ada-002`                 | No        | Hardcoded, not configurable        |
| `pgvector.WithPreDelete` | Always `true`                        | No        | Prevents accidental collection deletion |

## Design Notes
- **Hardcoded Model**: The embedding model is fixed to `text-embedding-ada-002` and cannot be changed via configuration.
- **Input Assumptions**: The package assumes valid inputs (non-empty URLs/tokens) and does not handle invalid parameters.
- **Error Handling**: Critical operations include explicit error handling.
- **No Dead Code**: The codebase is minimal and focused, with no TODOs or comments present.

## File Structure
- `store.go`: Contains the core implementation for vector store initialization and management.