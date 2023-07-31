# FireFly III Transaction Classification

This little web service performs transaction classification and integrates with FireFly III via web hooks. 

### How it works?
Every time you add new transaction to FireFly III, either manually or via import tool, web hook will trigger and provide transaction description to `fftc`. It will then be classified using Naive Bayesian Classification and transaction will be updated with matching category. Credits for Bayesian package goes to @navossoc.


