# Txt2PromQL

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/agentkube/txt2promql)](https://goreportcard.com/report/github.com/agentkube/txt2promql)
[![Build Status](https://github.com/agentkube/txt2promql/actions/workflows/go.yml/badge.svg)](https://github.com/agentkube/txt2promql/actions)

Convert natural language queries to PromQL with AI-powered understanding. Designed for monitoring democratization and observability workflows.

<!-- **Demo** (insert animated GIF here showing CLI and web interface) -->

## Features

- **Natural Language Interface**: Convert plain English to production-ready PromQL
- **Knowledge Graph Integration**: Understands metric relationships and monitoring best practices
- **Hybrid AI Approach**: Combines LLM capabilities with rule-based validation
- **Multi-Interface Support**:
  - REST API
  - Command Line Interface (CLI)
  - Web Dashboard
- **Prometheus Native**:
  - Auto schema discovery
  - Query validation
  - Built-in connection management

## Installation

### From Source

```bash
go install github.com/agentkube/txt2promql@latest
