# Overview

It's design to make SEO for single page application easier, so that you don't have to use Server-side rendering tricks.

It acts like a transparent http proxy by default, it only actives when the client looks like a web crawler, such as Googlebot, Baiduspider, etc.

## Installation

Usually, you will have a gateway like nginx in front of your web server. You can configure the gateway to proxy the request to bartender when the client looks like a web crawler.

A common way to detect web crawler: [link](https://stackoverflow.com/a/2517444/1089063).

A common data flow looks like this:

```mermaid
graph TD;
  C[Client]-->T[Gateway];
  T-->J{Is web crawler?};
  J-->|Yes|B[bartender];
  J-->|No|H[Your web server];
  B-->H;
```
