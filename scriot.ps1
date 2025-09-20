# PowerShell script to create Financial News Analysis project structure
# Run this script in the financial-news-analysis root directory

Write-Host "Creating Financial News Analysis project structure..." -ForegroundColor  Green

# Root files
$rootFiles = @(
    "README.md",
    "go.work",
    "docker-compose.yml",
    ".gitignore"
)

Write-Host "Creating root files..." -ForegroundColor  Yellow
foreach ($file in $rootFiles) {
    if (-not (Test-Path $file)) {
        New-Item -ItemType File -Path $file -Force | Out-Null
        Write-Host "  Created: $file" -ForegroundColor  Gray
    }
}

# Proto directory structure
Write-Host "Creating proto directory structure..." -ForegroundColor  Yellow
$protoStructure = @{
    "proto/article" = @("article.proto", "article_grpc.pb.go")
    "proto/analysis" = @("analysis.proto", "analysis_grpc.pb.go")
    "proto/market" = @("market.proto", "market_grpc.pb.go")
    "proto/signal" = @("signal.proto", "signal_grpc.pb.go")
    "proto/common" = @("common.proto", "common_grpc.pb.go")
}

foreach ($dir in $protoStructure.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $protoStructure[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

# Shared directory structure
Write-Host "Creating shared directory structure..." -ForegroundColor  Yellow
$sharedGoFiles = @{
    "shared/go/config" = @("config.go")
    "shared/go/logger" = @("logger.go")
    "shared/go/middleware" = @("auth.go", "tracing.go")
    "shared/go/utils" = @("grpc_client.go", "validation.go")
}

foreach ($dir in $sharedGoFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $sharedGoFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

$sharedJavaFiles = @{
    "shared/java/common/src/main/java/com/fintech/common/config" = @("GrpcConfig.java")
    "shared/java/common/src/main/java/com/fintech/common/exception" = @("ServiceException.java")
    "shared/java/common/src/main/java/com/fintech/common/util" = @("GrpcUtil.java")
    "shared/java/common" = @("pom.xml")
}

foreach ($dir in $sharedJavaFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $sharedJavaFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

# News Ingestion Service (Go)
Write-Host "Creating news-ingestion service structure..." -ForegroundColor  Yellow
$newsIngestionFiles = @{
    "services/news-ingestion" = @("go.mod", "go.sum", "main.go")
    "services/news-ingestion/config" = @("config.yaml")
    "services/news-ingestion/internal/handler" = @("grpc_handler.go", "http_handler.go")
    "services/news-ingestion/internal/service" = @("ingestion_service.go", "scraper_service.go", "deduplication_service.go")
    "services/news-ingestion/internal/repository" = @("article_repository.go", "cache_repository.go")
    "services/news-ingestion/internal/model" = @("article.go", "source.go")
    "services/news-ingestion/internal/client" = @("news_api_client.go", "rss_client.go", "twitter_client.go")
    "services/news-ingestion/pkg/scraper" = @("reuters.go", "bloomberg.go", "rss.go")
    "services/news-ingestion/pkg/validator" = @("article_validator.go")
    "services/news-ingestion/test/integration" = @("ingestion_test.go")
    "services/news-ingestion/test/unit" = @("service_test.go")
}

foreach ($dir in $newsIngestionFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $newsIngestionFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

# NLP Processing Service (Go)
Write-Host "Creating nlp-processing service structure..." -ForegroundColor  Yellow
$nlpProcessingFiles = @{
    "services/nlp-processing" = @("go.mod", "go.sum", "main.go")
    "services/nlp-processing/config" = @("config.yaml")
    "services/nlp-processing/internal/handler" = @("grpc_handler.go")
    "services/nlp-processing/internal/service" = @("nlp_service.go", "sentiment_service.go", "ner_service.go")
    "services/nlp-processing/internal/repository" = @("analysis_repository.go")
    "services/nlp-processing/internal/model" = @("analysis.go", "sentiment.go", "entity.go")
    "services/nlp-processing/pkg/ml/finbert" = @("model.go", "inference.go")
    "services/nlp-processing/pkg/ml/ner" = @("spacy_client.go", "entity_extractor.go")
    "services/nlp-processing/pkg/ml/topic" = @("classifier.go")
    "services/nlp-processing/pkg/preprocessing" = @("text_cleaner.go", "tokenizer.go")
    "services/nlp-processing/models/finbert" = @("model.bin")
    "services/nlp-processing/models/ner" = @("model.bin")
    "services/nlp-processing/models/topic" = @("classifier.bin")
    "services/nlp-processing/test/integration" = @("nlp_test.go")
    "services/nlp-processing/test/unit" = @("sentiment_test.go")
}

foreach ($dir in $nlpProcessingFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $nlpProcessingFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

# Scripts directory
Write-Host "Creating scripts directory..." -ForegroundColor  Yellow
$scriptFiles = @{
    "scripts" = @("generate-proto.sh", "start-local.sh", "setup-db.sh")
}

foreach ($dir in $scriptFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $scriptFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}

# Documentation directory
Write-Host "Creating docs directory..." -ForegroundColor  Yellow
$docFiles = @{
    "docs/api" = @("grpc-apis.md", "rest-apis.md")
    "docs" = @("architecture.md", "deployment.md", "development.md")
}

foreach ($dir in $docFiles.Keys) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    foreach ($file in $docFiles[$dir]) {
        $filePath = Join-Path $dir $file
        New-Item -ItemType File -Path $filePath -Force | Out-Null
        Write-Host "  Created: $filePath" -ForegroundColor  Gray
    }
}



