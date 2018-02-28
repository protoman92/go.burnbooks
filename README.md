# goburnbooks

[![Build Status](https://travis-ci.org/protoman92/goburnbooks.svg?branch=master)](https://travis-ci.org/protoman92/goburnbooks)

I was watching a presentation by **Rob Pike** (**Concurrency is not parallelism:** <https://vimeo.com/49718712>), and he mentioned the problem of book-burning gophers, whereby there are variable number of book piles/gophers/incinerators and the gophers are responsible for bringing books to the incinerators. Since I'm picking up Golang, might as well build some novice concurrency system just for fun.

The system comprises the following players:

- **SupplyPile**: Each pile has a number of **Suppliables**, and can provide those to interested **SupplyTakers**. A **SupplyPile** has to provide up to the capacity of a **SupplyTaker**, but there is always a risk of insufficient **Suppliables**.

- **Incinerator**: Each incinerator has a fixed capacity, and can consume **Burnables** provided by **BurnableProviders**. An **Incinerator** has to burn a certain load before it can signal that it is ready for more.

- **Gopher**: Each acts as both **SupplyTaker** and **BurnableProvider**, bringing **Books** from any **SupplyPile** to any **Incinerator**.

Each **Book** has a fixed burn duration, and each trip from a **SupplyPile** to an **Incinerator** may cost some time. **Gophers** need to wait for ready signals from **Incinerators** before they can start depositing **Books**, while **SupplyPiles** need to wait for ready signals from **Gophers** before they can begin supplying. The difficulty with this exercise lies with the chaotic interactions between many **SupplyPiles**, **Gophers** and **Incinerators** and the accurate processing of **Books** so that all are burned and unique.
