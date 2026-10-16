#!/usr/bin/env swift
// @(#)quiet / swift / ports/terminal/terminal-app/generate.swift / 2026-10-16

import AppKit
import Foundation

// Terminal.app serializes profile colors as NSKeyedArchiver-encoded NSColor
// objects embedded as Data values in an XML property list.
//
// This adapter deliberately uses AppKit rather than reproducing Cocoa archive
// internals in Go. Quiet Core owns color semantics; this program owns only the
// Terminal.app host representation.

private let profileVersion = 2.09

private let ansiKeys = [
    "ANSIBlackColor",
    "ANSIRedColor",
    "ANSIGreenColor",
    "ANSIYellowColor",
    "ANSIBlueColor",
    "ANSIMagentaColor",
    "ANSICyanColor",
    "ANSIWhiteColor",
    "ANSIBrightBlackColor",
    "ANSIBrightRedColor",
    "ANSIBrightGreenColor",
    "ANSIBrightYellowColor",
    "ANSIBrightBlueColor",
    "ANSIBrightMagentaColor",
    "ANSIBrightCyanColor",
    "ANSIBrightWhiteColor",
]

private struct ProfileDocument: Decodable {
    let name: String
    let background: String
    let text: String
    let bold: String
    let selection: String
    let cursor: String
    let ansi: [String]
}

private enum GeneratorError: Error, CustomStringConvertible {
    case usage
    case invalidHex(String)
    case invalidANSIColorCount(Int)
    case archiveFailed(String)
    case invalidPropertyList

    var description: String {
        switch self {
        case .usage:
            return "usage: generate.swift INPUT.json OUTPUT.terminal"

        case .invalidHex(let value):
            return "invalid #RRGGBB color \(String(reflecting: value))"

        case .invalidANSIColorCount(let count):
            return "expected 16 ANSI colors, got \(count)"

        case .archiveFailed(let key):
            return "could not archive NSColor for \(key)"

        case .invalidPropertyList:
            return "generated profile is not a valid property list"
        }
    }
}

private struct RGB {
    let red: CGFloat
    let green: CGFloat
    let blue: CGFloat

    init(hex: String) throws {
        guard
            hex.count == 7,
            hex.first == "#"
        else {
            throw GeneratorError.invalidHex(hex)
        }

        let digits = String(hex.dropFirst())

        guard
            let value = UInt32(digits, radix: 16)
        else {
            throw GeneratorError.invalidHex(hex)
        }

        red = CGFloat((value >> 16) & 0xff) / 255.0
        green = CGFloat((value >> 8) & 0xff) / 255.0
        blue = CGFloat(value & 0xff) / 255.0
    }

    func nsColor(alpha: CGFloat = 1.0) -> NSColor {
        NSColor(
            srgbRed: red,
            green: green,
            blue: blue,
            alpha: alpha
        )
    }
}

// archiveColor produces the representation Terminal.app expects for profile
// colors. Secure coding is intentionally not required because Terminal.app's
// exported profile format uses ordinary NSKeyedArchiver NSColor archives.
private func archiveColor(
    _ hex: String,
    key: String
) throws -> Data {
    let rgb = try RGB(hex: hex)

    do {
        return try NSKeyedArchiver.archivedData(
            withRootObject: rgb.nsColor(),
            requiringSecureCoding: false
        )
    } catch {
        throw GeneratorError.archiveFailed(key)
    }
}

private func buildProfile(
    from source: ProfileDocument
) throws -> [String: Any] {
    guard source.ansi.count == ansiKeys.count else {
        throw GeneratorError.invalidANSIColorCount(
            source.ansi.count
        )
    }

    var profile: [String: Any] = [
        "ProfileCurrentVersion": profileVersion,
        "name": source.name,
        "type": "Window Settings",
        "DisableANSIColor": false,
    ]

    // Quiet's resolved Selection color is an opaque display color. The
    // adapter therefore archives alpha=1.0 rather than inheriting
    // Terminal.app's profile-default selection opacity.
    profile["BackgroundColor"] = try archiveColor(
        source.background,
        key: "BackgroundColor"
    )

    profile["TextColor"] = try archiveColor(
        source.text,
        key: "TextColor"
    )

    profile["TextBoldColor"] = try archiveColor(
        source.bold,
        key: "TextBoldColor"
    )

    profile["SelectionColor"] = try archiveColor(
        source.selection,
        key: "SelectionColor"
    )

    profile["CursorColor"] = try archiveColor(
        source.cursor,
        key: "CursorColor"
    )

    for (index, key) in ansiKeys.enumerated() {
        profile[key] = try archiveColor(
            source.ansi[index],
            key: key
        )
    }

    return profile
}

private func run() throws {
    guard CommandLine.arguments.count == 3 else {
        throw GeneratorError.usage
    }

    let inputURL = URL(
        fileURLWithPath: CommandLine.arguments[1]
    )

    let outputURL = URL(
        fileURLWithPath: CommandLine.arguments[2]
    )

    let sourceData = try Data(
        contentsOf: inputURL
    )

    let source = try JSONDecoder().decode(
        ProfileDocument.self,
        from: sourceData
    )

    let profile = try buildProfile(
        from: source
    )

    guard PropertyListSerialization.propertyList(
        profile,
        isValidFor: .xml
    ) else {
        throw GeneratorError.invalidPropertyList
    }

    let plist = try PropertyListSerialization.data(
        fromPropertyList: profile,
        format: .xml,
        options: 0
    )

    try plist.write(
        to: outputURL,
        options: .atomic
    )
}

do {
    try run()
} catch {
    FileHandle.standardError.write(
        Data(
            "quiet-terminal-app: \(error)\n".utf8
        )
    )

    exit(1)
}

