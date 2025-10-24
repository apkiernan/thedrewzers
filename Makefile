# Development targets
.PHONY: all tpl styles server build clean deploy lambda-build upload-static tf-init tf-plan tf-apply tf-destroy static-build static-deploy invalidate-cache gallery-metadata optimize-images

# Default development target
all: dev-build
	@echo "Starting development servers..."
	@make -j2 styles server

# Prepare dist directory for development
dev-build: tpl gallery-metadata optimize-images
	@echo "Copying static assets to dist..."
	@mkdir -p dist
	@cp -r static/css static/js static/fonts static/data dist/ 2>/dev/null || true
	@cp static/gallery-metadata.json dist/ 2>/dev/null || true
	@echo "Development build complete!"

# Build local development server
build:
	go build -o bin/main ./cmd/main

# Run local development server with hot reload
server:
	air

# Generate templ templates
tpl:
	templ generate

# Watch and compile styles
styles:
	npm run watch

# Generate gallery metadata
gallery-metadata:
	@echo "Generating gallery image metadata..."
	@go run cmd/gallery-metadata/main.go

# Image optimization variables
SOURCE_IMAGES=static/images
DIST_IMAGES=dist/images

# Optimize images: generate responsive sizes and modern formats
optimize-images:
	@echo "Creating output directory..."
	@mkdir -p $(DIST_IMAGES)
	@echo "Generating responsive image sizes..."
	@go run cmd/optimize-images/main.go
	@echo "Converting to WebP format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" ! -name "*-lqip.jpg" -exec sh -c \
		'cwebp -q 85 -preset photo -m 6 "$$1" -o "$${1%.jpg}.webp" 2>/dev/null && echo "  Converted: $$(basename $${1%.jpg}.webp)"' sh {} \;
	@echo "Converting to AVIF format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" ! -name "*-lqip.jpg" -exec sh -c \
		'avifenc --min 0 --max 63 -a end-usage=q -a cq-level=18 -a tune=ssim --jobs 8 "$$1" "$${1%.jpg}.avif" 2>/dev/null && echo "  Converted: $$(basename $${1%.jpg}.avif)"' sh {} \;
	@echo "Generating LQIP placeholders..."
	@go run cmd/generate-lqip/main.go
	@echo "Image optimization complete!"

# Build static HTML files
static-build: tpl optimize-images gallery-metadata
	@echo "Building static site..."
	@go run ./cmd/build
	@echo "Copying static assets to dist..."
	@cp -r static/* dist/
	@echo "Static site ready in ./dist/"

# Lambda Configuration
LAMBDA_BINARY=bootstrap
BUILD_DIR=build
LAMBDA_ZIP=$(BUILD_DIR)/lambda.zip
GO_BUILD_FLAGS=-ldflags="-s -w"

# AWS region and S3 bucket for static assets
AWS_REGION=us-east-1
S3_BUCKET=thedrewzers-wedding-static

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Build the Lambda binary
lambda-build: clean
	GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(LAMBDA_BINARY) ./cmd/lambda
	cd $(BUILD_DIR) && zip -r lambda.zip $(LAMBDA_BINARY)

# Upload static assets to S3
upload-static:
	aws s3 sync ./static s3://$(S3_BUCKET)/static/ --acl public-read

# Deploy static site to S3
static-deploy: static-build
	@echo "Uploading HTML files to S3..."
	@aws s3 cp dist/index.html s3://$(S3_BUCKET)/index.html --acl public-read --content-type "text/html"
	@aws s3 cp dist/venue.html s3://$(S3_BUCKET)/venue.html --acl public-read --content-type "text/html"
	@aws s3 cp dist/gallery.html s3://$(S3_BUCKET)/gallery.html --acl public-read --content-type "text/html"
	@echo "Uploading static assets to S3..."
	@aws s3 sync dist/css s3://$(S3_BUCKET)/static/css/ --acl public-read
	@aws s3 sync dist/js s3://$(S3_BUCKET)/static/js/ --acl public-read
	@aws s3 sync dist/fonts s3://$(S3_BUCKET)/static/fonts/ --acl public-read
	@aws s3 sync dist/images s3://$(S3_BUCKET)/static/images/ --acl public-read
	@echo "Static site deployed successfully!"
	@echo "Creating CloudFront invalidation..."
	@if [ -n "$(CLOUDFRONT_DISTRIBUTION_ID)" ]; then \
		aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_DISTRIBUTION_ID) --paths "/*" --query 'Invalidation.Id' --output text; \
		echo "CloudFront invalidation created successfully!"; \
	else \
		DISTRIBUTION_ID=$$(cd terraform && terraform output -raw cloudfront_distribution_id 2>/dev/null) && \
		if [ -n "$$DISTRIBUTION_ID" ] && [ "$$DISTRIBUTION_ID" != "" ]; then \
			aws cloudfront create-invalidation --distribution-id $$DISTRIBUTION_ID --paths "/*" --query 'Invalidation.Id' --output text; \
			echo "CloudFront invalidation created successfully!"; \
		else \
			echo "Warning: CloudFront distribution ID not found. Skipping invalidation."; \
			echo "You can set CLOUDFRONT_DISTRIBUTION_ID environment variable or run 'terraform apply' first."; \
		fi \
	fi

# Invalidate CloudFront cache
invalidate-cache:
	@echo "Creating CloudFront invalidation..."
	@if [ -n "$(CLOUDFRONT_DISTRIBUTION_ID)" ]; then \
		aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_DISTRIBUTION_ID) --paths "/*" --query 'Invalidation.Id' --output text; \
		echo "CloudFront invalidation created successfully!"; \
	else \
		DISTRIBUTION_ID=$$(cd terraform && terraform output -raw cloudfront_distribution_id 2>/dev/null) && \
		if [ -n "$$DISTRIBUTION_ID" ] && [ "$$DISTRIBUTION_ID" != "" ]; then \
			aws cloudfront create-invalidation --distribution-id $$DISTRIBUTION_ID --paths "/*" --query 'Invalidation.Id' --output text; \
			echo "CloudFront invalidation created successfully!"; \
		else \
			echo "Error: CloudFront distribution ID not found."; \
			echo "Set CLOUDFRONT_DISTRIBUTION_ID environment variable or run 'terraform apply' first."; \
			exit 1; \
		fi \
	fi

# Deploy with Terraform
deploy: gallery-metadata lambda-build static-deploy
	cd terraform && terraform init
	cd terraform && terraform apply

# Destroy all resources
tf-destroy:
	cd terraform && terraform destroy

# Initialize Terraform
tf-init:
	cd terraform && terraform init

# Plan Terraform changes
tf-plan:
	cd terraform && terraform plan

# Apply Terraform changes
tf-apply:
	cd terraform && terraform apply

