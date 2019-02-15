const SimpleToken = artifacts.require("SimpleToken");
const Crowdsale = artifacts.require("Crowdsale");


// this.token = await SimpleToken.new();
// this.crowdsale = await Crowdsale.new(rate, wallet, this.token.address);
// await this.token.transfer(this.crowdsale.address, tokenSupply);

contract('Crowdsale test', async (accounts) => {
  it("blabla", async () => {
    let token = await SimpleToken.new();
    let crowdsale = await Crowdsale.new(new web3.BigNumber(1), web3.eth.accounts[0], token.address);
    await token.transfer(crowdsale.address, new web3.BigNumber('1e27'));
    await crowdsale.buyTokens(web3.eth.accounts[0], {value:5});
    await crowdsale.buyTokens(web3.eth.accounts[1], {value:7});

    let balance0 = await token.balanceOf(web3.eth.accounts[0]);
    let balance1 = await token.balanceOf(web3.eth.accounts[1]);
    console.log(balance0.toNumber())
    console.log(balance1.toNumber())
  })
})
