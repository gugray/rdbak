# Raindrop bookmark backup

This is a minimalistic command line tool that downloads a local backup of your [Raindrop.io](https://raindrop.io/) bookmarks.

- Raindrop keeps a permanent copy of your bookmarks in the cloud. I.e., it stores not only the URL and page title, but the full content.
- This tool retrieves your bookmarks and saves their permanent copy in your local file system.
- Raindrop's permanent copies are self-contained HTML files, which is what this tool will download for you.
- The tool works incrementally: it saves a JSON with your bookmarks, and only downloads new and changed permanent copies when you run it again.
- Downloaded files are never deleted.
- The tool can be built to run on low-end Synology disk stations with ARMv7 processors.

### Why did I build this?

Raindrop's offline copies are the best thing since sliced bread! They enable full-text search in the pages you've bookmarked, and they stick around even if the original page disappears.
But they are still in the cloud somewhere, and I'd lose access to them if I ever closed my account, or if Raindrop went out of business.
So given that Raindrop has already done the hard work of fetching the pages and storing them as self-contained HTML files, it felt like a natural step to maintain an offline archive for myself.

### Is this legit?

Yes.

The tool uses Raindrop's private API - the same one through which the web app talks to the backend. A service's developer is free to change this kind of API at any time,
even in ways that break compatibility. It's not likely to happen often, but it's entirely possible. It would mean that `rdbak` stops working in its current form.

I mailed Raindrop's awesome maintainer, Rustem Mussabekov, if he was OK with me creating this tool and sharing it publicly. He responded that this was fine, but emphasized
the caveat about the private API.

## Build and run from source

Make a copy of `config.sample.json` named `config.dev.json` and edit the file. Add the user name and password you use to log in to Raindrop.io at [app.raindrop.ip](app.raindrop.io)
or via the browser addin.

Build and run `rdbak` from source with this command from the repository root:

```
go run . encrypt-pwd
```

This replaces your plaintext password in the config file with an encrypted string, using a hard-wired key. You should still not share the config file with the encrypted password
in it, but it's a minimal level of protection against your password getting stolen if someone gains access to your computer.

Subsequently, you can download your new bookmarks by running:

```
go run . backup
```

