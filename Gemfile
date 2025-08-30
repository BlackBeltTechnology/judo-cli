source "https://rubygems.org"

# Use Ruby 2.6 compatible versions
ruby ">= 2.6.0"

# Jekyll with compatible version (downgrade to avoid eventmachine)
gem "jekyll", "~> 4.0.0"
gem "webrick", "~> 1.7"

# Theme dependencies
gem "kramdown-parser-gfm"
gem "rouge", "~> 3.0"

# Group all Jekyll plugins together (minimal set to avoid native compilation)
group :jekyll_plugins do
  gem "jekyll-remote-theme"
  gem "jekyll-sitemap" 
  gem "jekyll-seo-tag"
end