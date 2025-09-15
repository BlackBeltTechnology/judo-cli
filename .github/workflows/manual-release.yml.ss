name: Manual Release

# Default environment variables for better error handling
env:
  CI: true
  MAX_RETRIES: 3
  RETRY_DELAY: 10

on:
  workflow_dispatch:
    inputs:
      version_type:
        description: 'Version increment type'
        required: true
        default: 'patch'
        type: choice
        options:
          - patch
          - minor
          - major
      custom_version:
        description: 'Custom version (optional, overrides version_type)'
        required: false
        type: string

jobs:
  create-release:
    name: Create Manual Release
    runs-on: judong
    # Continue on error to ensure proper cleanup
    continue-on-error: true
    steps:
      - name: Checkout develop branch with full history
        uses: actions/checkout@v4
        with:
          ref: develop
          fetch-depth: 0
          fetch-tags: true
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Configure Git
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"

      - name: Determine new version with validation
        id: new_version
        run: |
          set -e
          echo "Determining new version..."
          
          # Validate version.sh script
          if [[ ! -f "scripts/version.sh" ]]; then
            echo "::error::version.sh script not found"
            exit 1
          fi
          
          chmod +x scripts/version.sh
          
          # Test the script works
          if ! scripts/version.sh get > /dev/null 2>&1; then
            echo "::error::version.sh script validation failed"
            exit 1
          fi
          
          if [[ -n "${{ github.event.inputs.custom_version }}" ]]; then
            NEW_VERSION="${{ github.event.inputs.custom_version }}"
            echo "Using custom version: $NEW_VERSION"
            
            # Validate custom version format
            if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
              echo "::error::Invalid custom version format: $NEW_VERSION. Expected X.Y.Z"
              exit 1
            fi
          else
            VERSION_TYPE="${{ github.event.inputs.version_type }}"
            echo "Incrementing $VERSION_TYPE version..."
            
            # Validate version type
            if [[ ! "$VERSION_TYPE" =~ ^(patch|minor|major)$ ]]; then
              echo "::error::Invalid version type: $VERSION_TYPE. Must be patch, minor, or major"
              exit 1
            fi
            
            NEW_VERSION=$(scripts/version.sh increment "$VERSION_TYPE")
            echo "Incremented $VERSION_TYPE version: $NEW_VERSION"
          fi
          
          # Validate the new version
          if [[ -z "$NEW_VERSION" ]]; then
            echo "::error::Failed to determine new version"
            exit 1
          fi
          
          if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "::error::Invalid new version format: $NEW_VERSION. Expected X.Y.Z"
            exit 1
          fi
          
          echo "new_version=${NEW_VERSION}" >> $GITHUB_OUTPUT
          echo "tag_name=v${NEW_VERSION}" >> $GITHUB_OUTPUT
          echo "✅ New version determined: $NEW_VERSION"

      - name: Update VERSION file and commit with validation
        run: |
          set -e
          NEW_VERSION="${{ steps.new_version.outputs.new_version }}"
          
          echo "Updating VERSION file to: $NEW_VERSION"
          
          # Update version file
          scripts/version.sh set "$NEW_VERSION" --commit
          
          # Verify the update worked
          UPDATED_VERSION=$(scripts/version.sh get)
          if [[ "$UPDATED_VERSION" != "$NEW_VERSION" ]]; then
            echo "::error::Failed to update VERSION file. Expected: $NEW_VERSION, Got: $UPDATED_VERSION"
            exit 1
          fi
          
          echo "✅ VERSION file updated and committed: $NEW_VERSION"

      - name: Push version update and tag with retry logic
        run: |
          set -e
          TAG_NAME="${{ steps.new_version.outputs.tag_name }}"
          
          echo "Pushing version update and tag..."
          
          # Verify we're on develop branch
          CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
          if [[ "$CURRENT_BRANCH" != "develop" ]]; then
            echo "::error::Expected to be on develop branch, but currently on: $CURRENT_BRANCH"
            exit 1
          fi
          
          # Push commit with retry logic
          for i in $(seq 1 ${{ env.MAX_RETRIES }}); do
            echo "Push commit attempt $i/${{ env.MAX_RETRIES }}"
            if git push origin develop; then
              echo "✅ Successfully pushed commit to develop branch"
              break
            else
              echo "::warning::Push failed (attempt $i/${{ env.MAX_RETRIES }}), retrying in ${{ env.RETRY_DELAY }} seconds..."
              if [[ $i -eq ${{ env.MAX_RETRIES }} ]]; then
                echo "::error::Failed to push to develop branch after ${{ env.MAX_RETRIES }} attempts"
                exit 1
              fi
              sleep ${{ env.RETRY_DELAY }}
            fi
          done
          
          # Create and push tag with retry logic
          echo "Creating and pushing tag: $TAG_NAME"
          git tag "$TAG_NAME"
          
          for i in $(seq 1 ${{ env.MAX_RETRIES }}); do
            echo "Push tag attempt $i/${{ env.MAX_RETRIES }}"
            if git push origin "$TAG_NAME"; then
              echo "✅ Successfully pushed tag $TAG_NAME"
              break
            else
              echo "::warning::Tag push failed (attempt $i/${{ env.MAX_RETRIES }}), retrying in ${{ env.RETRY_DELAY }} seconds..."
              if [[ $i -eq ${{ env.MAX_RETRIES }} ]]; then
                echo "::error::Failed to push tag after ${{ env.MAX_RETRIES }} attempts"
                exit 1
              fi
              sleep ${{ env.RETRY_DELAY }}
            fi
          done

      - name: Create comprehensive summary
        run: |
          set -e
          echo "## Manual Release Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          
          if [[ "${{ job.status }}" == "success" ]]; then
            echo "✅ **Status**: SUCCESS" >> $GITHUB_STEP_SUMMARY
          else
            echo "❌ **Status**: FAILED" >> $GITHUB_STEP_SUMMARY
          fi
          
          echo "- **Version Type**: ${{ github.event.inputs.version_type }}" >> $GITHUB_STEP_SUMMARY
          echo "- **New Version**: ${{ steps.new_version.outputs.new_version }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Release Tag**: ${{ steps.new_version.outputs.tag_name }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Branch**: develop" >> $GITHUB_STEP_SUMMARY
          
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "### Next Steps" >> $GITHUB_STEP_SUMMARY
          if [[ "${{ job.status }}" == "success" ]]; then
            echo "- The release workflow will be automatically triggered by the tag push" >> $GITHUB_STEP_SUMMARY
            echo "- Monitor the release progress [here](${{ github.server_url }}/${{ github.repository }}/actions)" >> $GITHUB_STEP_SUMMARY
          else
            echo "- Review the error logs above to identify the issue" >> $GITHUB_STEP_SUMMARY
            echo "- Check that all required files and permissions are in place" >> $GITHUB_STEP_SUMMARY
            echo "- Retry the manual release after addressing the issues" >> $GITHUB_STEP_SUMMARY
          fi
          
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "### Workflow Details" >> $GITHUB_STEP_SUMMARY
          echo "- **Workflow Run**: [#${{ github.run_number }}](${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }})" >> $GITHUB_STEP_SUMMARY
          echo "- **Triggered By**: ${{ github.actor }}" >> $GITHUB_STEP_SUMMARY
          
          # Set final exit code based on job status
          if [[ "${{ job.status }}" != "success" ]]; then
            echo "::error::Manual release workflow failed"
            exit 1
          fi