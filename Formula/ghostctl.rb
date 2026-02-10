class Ghostctl < Formula
  desc "CLI tool for managing ephemeral Kubernetes clusters using vCluster"
  homepage "https://github.com/ghostcluster-ai/ghostctl"
  url "https://github.com/ghostcluster-ai/ghostctl.git",
      branch: "main"
  version "1.0.0"
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    # Build the binary
    system "make", "build"
    
    # Install to bin directory
    bin.install "bin/ghostctl"
  end

  test do
    system "#{bin}/ghostctl", "--help"
  end
end
