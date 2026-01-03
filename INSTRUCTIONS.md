I would like your help building a brand new stock checker app. To start we are only concerned with running this app locally for myself and getting basic functionality working. After this we can work on getting persistent storage + deploying it out.

Here is the problem I am solving. Pokemon cards are currently in very high demand and often go out of stock very very quickly at popular stores such as Best Buy, Target, Costco, Sams Club, Walmart, etc.

I would like your help building an app that can check for new pokemon products coming into stores near me so that I can be notified when the stores closest to me get new inventory and I can then go into the store to buy them.

I know that Best Buy has an API which can be used, so lets start with just this store for our app.

Here is the functionality we should build
1. Start a Frontend app locally
2. Be able to search  for best buy stores
3. Be able to add specific stores to "My Stores List"
4. Be able to remove stores from "My Stores List"
5. Be able to search for best buy products by name and product number
6. Be able to add specific products to "My Products List"
7. Be able to remove stores from "My Products List"
8. There should be a button to "Check Stock", which looks for the products in "My Products List" and see's if there is inventory in "My Stores List"
9. Provides a nice easy table to read through the results that makes it clear what store has inventory of what product, sorted by the products that have inventory first.

I am working on getting us the API Key from BestBuy, so lets stub that for now and use an EnvVar.

I would like the FE to be in React + Typescript. I would like the backend to use protobufs + golang. Is that possible? I would like you to write up a plan in PLAN.md with all of the various tasks we will need to complete to achieve and build this app so we can keep track. Lets keep track of the features we want to implement as well in case we want to add more in the future. Do you have any questions about how we should implement this?
