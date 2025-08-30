# This is a template showing what the auto-generated Homebrew formula will look like
# The actual formula will be generated automatically by GoReleaser in the homebrew-tap repository

class Judo < Formula
  desc "JUDO CLI - A command-line tool for managing the lifecycle of JUDO applications"
  homepage "https://github.com/BlackBeltTechnology/judo-cli"
  license "EPL-2.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/BlackBeltTechnology/judo-cli/releases/download/v1.0.0/judo_Darwin_x86_64.tar.gz"
      sha256 "CHECKSUM_WILL_BE_AUTO_GENERATED"
    end
    if Hardware::CPU.arm?
      url "https://github.com/BlackBeltTechnology/judo-cli/releases/download/v1.0.0/judo_Darwin_arm64.tar.gz"
      sha256 "CHECKSUM_WILL_BE_AUTO_GENERATED"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/BlackBeltTechnology/judo-cli/releases/download/v1.0.0/judo_Linux_x86_64.tar.gz"
      sha256 "CHECKSUM_WILL_BE_AUTO_GENERATED"
    end
    if Hardware::CPU.arm?
      url "https://github.com/BlackBeltTechnology/judo-cli/releases/download/v1.0.0/judo_Linux_arm64.tar.gz"
      sha256 "CHECKSUM_WILL_BE_AUTO_GENERATED"
    end
  end

  depends_on "docker" => :optional
  depends_on "maven" => :optional

  def install
    bin.install "judo"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/judo version")
  end
end