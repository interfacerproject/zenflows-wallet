<!--
SPDX-License-Identifier: AGPL-3.0-or-later
Copyright (C) 2023 Dyne.org foundation <foundation@dyne.org>.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

<div align="center">

# Zenflows wallet

### Zenflows economic model API's

</div>

<p align="center">
  <a href="https://dyne.org">
    <img src="https://files.dyne.org/software_by_dyne.png" width="170">
  </a>
</p>

## Building the digital infrastructure for Fab Cities

<br>
<a href="https://www.interfacerproject.eu/">
  <img alt="Interfacer project" src="https://dyne.org/images/projects/Interfacer_logo_color.png" width="350" />
</a>
<br>

### What is **INTERFACER?**

The goal of the INTERFACER project is to build the open-source digital infrastructure for Fab Cities.

Our vision is to promote a green, resilient, and digitally-based mode of production and consumption that enables the greatest possible sovereignty, empowerment and participation of citizens all over the world.
We want to help Fab Cities to produce everything they consume by 2054 on the basis of collaboratively developed and globally shared data in the commons.

To know more [DOWNLOAD THE WHITEPAPER](https://www.interfacerproject.eu/assets/news/whitepaper/IF-WhitePaper_DigitalInfrastructureForFabCities.pdf)

## Zenflows wallet Features

* Economic model based on Zenflows actions
* Stored on federated Tarantool distributed storage
* Solid crypto based on [zenflows-crypto](https://github.com/interfacerproject/zenflows-crypto)

# [LIVE DEMO](https://interfacer-gui-staging.dyne.org/)

<br>

<div id="toc">

### 🚩 Table of Contents

- [🎮 Quick start](#-quick-start)
- [🐋 Docker](#-docker)
- [🐝 API](#-api)
- [🔧 Build](#-build)
- [😍 Acknowledgements](#-acknowledgements)
- [🌐 Links](#-links)
- [👤 Contributing](#-contributing)
- [💼 License](#-license)

</div>


***
## 🎮 Quick start

To start using Zenflows wallet just run in the root folder

```bash
docker compose up
```

**[🔝 back to top](#toc)**

***
## 🐋 Docker

Container are published at each commit on main via Github Actions

```bash
docker pull ghcr.io/interfacerproject/zenflows-wallet:main
docker pull ghcr.io/interfacerproject/zenflows-wallet-db:main
```

**[🔝 back to top](#toc)**

***
## 🐝 API

Endpoints are available

### POST /token

Execute a transaction with some amount

**Parameters**

|          Name | Required |  Type   | Description       | 
| -------------:|:--------:|:-------:| ------------------|
|       `token` | required | string  | Type of token. Accepted values `idea` or `strength`  |
|       `amount`| required | number  | Transaction's token amount |
|       `owner` | required | ULID    | The ULID of the Agent's owner |
 
### GET /token/${request.token}/${request.owner}

Retrieves the actual value of the token type for the specified owner


**[🔝 back to top](#toc)**

***
## 🔧 Build

```bash
make
```

**[🔝 back to top](#toc)**

***
## 😍 Acknowledgements

<a href="https://dyne.org">
  <img src="https://files.dyne.org/software_by_dyne.png" width="222">
</a>

Copyleft (ɔ) 2022 by [Dyne.org](https://www.dyne.org) foundation, Amsterdam

Designed, written and maintained by Alberto Lerda

With contributions of Puria Nafisi Azizi

### Funding

#### October 2021 – March 2023

This project was funded by the European Regional Development Fund (ERDF) as part of the INTERFACER project.

<img src="docs/assets/funding/logo_erdf2014.png" alt="EU ERDF Programme Logo" width="480">

#### January 2025 – December 2026

The further development of this project was supported as part of the "Fab City – Decentralized Digital Production for Urban Value Creation" project, funded by dtec.bw and supported by the European Union – NextGenerationEU.

<p>
  <img src="docs/assets/funding/dtec.bw_Logo_RGB_komplett.jpg" alt="dtec.bw logo" width="240">
  <img src="docs/assets/funding/nextgeneu_en.jpg" alt="Supported by the European Union – NextGenerationEU" width="360">
</p>

**[🔝 back to top](#toc)**

***
## 🌐 Links

https://www.interfacer.eu/

https://dyne.org/

**[🔝 back to top](#toc)**

***
## 👤 Contributing

1.  🔀 [FORK IT](../../fork)
2.  Create your feature branch `git checkout -b feature/branch`
3.  Commit your changes `git commit -am 'Add some fooBar'`
4.  Push to the branch `git push origin feature/branch`
5.  Create a new Pull Request
6.  🙏 Thank you


**[🔝 back to top](#toc)**

***
## 💼 License
    Zenflows wallet - Zenflows economic model API's
    Copyleft (ɔ) 2022 Dyne.org foundation

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

**[🔝 back to top](#toc)**
