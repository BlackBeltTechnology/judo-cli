# This Gemfile is for CI/CD only - local development may require proper Ruby environment
# For local development, the documentation is available at: https://judo.technology/

source "https://rubygems.org"

# Use Ruby 2.6 compatible versions
ruby ">= 2.6.0"

# Jekyll with compatible version for CI
gem "jekyll", "~> 4.2.0"
gem "webrick", "~> 1.7"

# Theme dependencies
gem "kramdown-parser-gfm"
gem "rouge", "~> 3.0"

# Jekyll plugins (required for GitHub Pages build)
group :jekyll_plugins do
  gem "jekyll-remote-theme"
  gem "jekyll-feed"
  gem "jekyll-sitemap" 
  gem "jekyll-seo-tag"
end

# Development gems
group :development do
  gem "jekyll-watch", "~> 2.2"
end