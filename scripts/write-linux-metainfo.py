#!/usr/bin/env python3
"""Write the AppStream metainfo file installed with the Linux packages.

Opening a .deb in GNOME Software reads Homepage and License from the
package control data. After install, Software reads this metainfo file
for the license, the project website, and the release notes.
"""

import re
import sys
from datetime import date
from pathlib import Path
from xml.sax.saxutils import escape

ROOT = Path(__file__).resolve().parent.parent


def read_value(path, pattern, label):
    text = path.read_text(encoding="utf-8")
    match = re.search(pattern, text)
    if not match or not match.group(1).strip():
        raise SystemExit(f"error: {label} must be set in {path.relative_to(ROOT)}")
    value = match.group(1).strip()
    if "\n" in value or "\r" in value:
        raise SystemExit(f"error: {label} must be a single line")
    return value


def xml_text(value):
    return escape(value, {'"': "&quot;", "'": "&apos;"})


def summary_line(description):
    line = description.strip()
    if line.endswith("."):
        line = line[:-1].rstrip()
    if not line:
        raise SystemExit("error: info.description must contain a summary")
    return line


def developer_id(product_id):
    parts = product_id.split(".")
    if len(parts) >= 3:
        return ".".join(parts[:-1])
    return product_id


def categories_for(app_name):
    desktop = ROOT / "build" / "linux" / f"{app_name}.desktop"
    if desktop.exists():
        for line in desktop.read_text(encoding="utf-8").splitlines():
            if line.startswith("Categories="):
                names = [part for part in line.split("=", 1)[1].split(";") if part]
                if names:
                    return names
    # Matches the Development category written by linux:generate:dotdesktop.
    return ["Development"]


def main():
    config = ROOT / "build" / "config.yml"
    taskfile = ROOT / "Taskfile.yml"
    app_name = read_value(taskfile, r'(?m)^  APP_NAME: "([^"]+)"', "APP_NAME")
    product_name = read_value(config, r'(?m)^  productName: "([^"]+)"', "info.productName")
    product_id = read_value(config, r'(?m)^  productIdentifier: "([^"]+)"', "info.productIdentifier")
    description = read_value(config, r'(?m)^  description: "([^"]+)"', "info.description")
    company = read_value(config, r'(?m)^  companyName: "([^"]+)"', "info.companyName")
    version = read_value(config, r'(?m)^  version: "([^"]+)"', "info.version")
    homepage = read_value(config, r'(?m)^  homepage: "([^"]+)"', "info.homepage")
    license_id = read_value(config, r'(?m)^  license: "([^"]+)"', "info.license")

    if not re.fullmatch(r"[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+", product_id):
        raise SystemExit("error: info.productIdentifier must be a reverse-DNS id, such as com.example.app")
    if not homepage.startswith(("https://", "http://")):
        raise SystemExit("error: info.homepage must be an http or https URL")
    if not re.fullmatch(r"\d+\.\d+\.\d+", version):
        raise SystemExit("error: info.version must be X.Y.Z")
    if re.search(r"[<>&]", license_id):
        raise SystemExit("error: info.license must be an SPDX identifier, such as MIT")

    summary = summary_line(description)
    released = date.today().isoformat()
    categories = "\n".join(f"    <category>{xml_text(name)}</category>" for name in categories_for(app_name))
    body = f"""<?xml version="1.0" encoding="UTF-8"?>
<!-- Generated from build/config.yml by scripts/write-linux-metainfo.py. -->
<component type="desktop-application">
  <id>{xml_text(product_id)}</id>
  <metadata_license>{xml_text(license_id)}</metadata_license>
  <project_license>{xml_text(license_id)}</project_license>
  <name>{xml_text(product_name)}</name>
  <summary>{xml_text(summary)}</summary>
  <developer id="{xml_text(developer_id(product_id))}">
    <name>{xml_text(company)}</name>
  </developer>
  <description>
    <p>{xml_text(description)}</p>
  </description>
  <launchable type="desktop-id">{xml_text(app_name)}.desktop</launchable>
  <icon type="stock">{xml_text(app_name)}</icon>
  <categories>
{categories}
  </categories>
  <url type="homepage">{xml_text(homepage)}</url>
  <provides>
    <binary>{xml_text(app_name)}</binary>
  </provides>
  <releases>
    <release version="{xml_text(version)}" date="{released}">
      <description>
        <p>{xml_text(description)}</p>
      </description>
    </release>
  </releases>
  <content_rating type="oars-1.1"/>
</component>
"""
    dest = ROOT / "build" / "linux" / f"{app_name}.metainfo.xml"
    dest.write_text(body, encoding="utf-8")
    print(dest.relative_to(ROOT))


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        sys.exit(0)
