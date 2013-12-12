# Maintainer: ushi <ushi@honkgong.info>
pkgname='webmpc-git'
pkgver='0.0.0'
pkgrel=1
pkgdesc='realtime mpd client running in the browser'
arch=('x86_64' 'i686' 'armv6h')
url='https://put.honkgong.info'
license=('MIT')
conflicts=('webmpc')
provides=('webmpc')
makedepends=('go' 'ruby')
options=(!strip)
source=('webmpc::git+https://github.com/ushis/webmpc#branch=master')
sha256sums=('SKIP')

pkgver() {
  cd webmpc
  echo "$(git rev-list --count master).$(git rev-parse --short master)"
}

build() {
  cd webmpc
  make
}

package() {
  cd webmpc
  install -Dm0755 'webmpcd'                 "${pkgdir}/usr/bin/webmpcd"
  install -Dm0644 'html/index.html'         "${pkgdir}/usr/share/webmpc/index.html"
  install -Dm0644 'html/webmpc.css'         "${pkgdir}/usr/share/webmpc/webmpc.css"
  install -Dm0644 'html/webmpc.js'          "${pkgdir}/usr/share/webmpc/webmpc.js"
  install -Dm0644 'html/pause.png'          "${pkgdir}/usr/share/webmpc/pause.png"
  install -Dm0644 'html/play.png'           "${pkgdir}/usr/share/webmpc/play.png"
  install -Dm0644 'systemd/webmpcd.conf'    "${pkgdir}/usr/lib/tmpfiles.d/webmpcd.conf"
  install -Dm0644 'systemd/webmpcd.service' "${pkgdir}/usr/lib/systemd/system/webmpcd.service"
}

# vim:set ts=2 sw=2 et:
