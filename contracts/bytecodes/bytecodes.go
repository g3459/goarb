package bytecodes

import (
	"encoding/hex"
)

var RouterBytecode, _ = hex.DecodeString("6080604052348015600f57600080fd5b506111468061001f6000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80631cc0285e14610030575b600080fd5b61004361003e366004610d20565b610059565b6040516100509190610e0b565b60405180910390f35b60608367ffffffffffffffff81111561007457610074610edc565b6040519080825280602002602001820160405280156100ba57816020015b6040805180820190915260008152606060208201528152602001906001900390816100925790505b509050868187815181106100d0576100d0610f0b565b60209081029190910101515260606100ed898787878786886100f9565b50979650505050505050565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6101008690031c5b60005b86811015610b3b576001811b821615610b3357806001901b82189150600083828151811061015557610155610f0b565b60200260200101516000015190506000811115610b31578915610509577fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8a01600182901c60008a67ffffffffffffffff8111156101b5576101b5610edc565b6040519080825280602002602001820160405280156101e857816020015b60608152602001906001900390816101d35790505b50905060008b67ffffffffffffffff81111561020657610206610edc565b60405190808252806020026020018201604052801561024c57816020015b6040805180820190915260008152606060208201528152602001906001900390816102245790505b5090508281878151811061026257610262610f0b565b60209081029190910101515260005b8c8110156102bc5788818151811061028b5761028b610f0b565b6020026020010151602001518382815181106102a9576102a9610f0b565b6020908102919091010152600101610271565b506102cc848e8e8e8e87876100f9565b60005b8c811015610358578881815181106102e9576102e9610f0b565b60200260200101516020015182828151811061030757610307610f0b565b602002602001015160200151604051602001610324929190610f3a565b60405160208183030381529060405283828151811061034557610345610f0b565b60209081029190910101526001016102cf565b5060008c67ffffffffffffffff81111561037457610374610edc565b6040519080825280602002602001820160405280156103ba57816020015b6040805180820190915260008152606060208201528152602001906001900390816103925790505b509050838188815181106103d0576103d0610f0b565b6020908102919091010151526103eb858f8f8f8f88876100f9565b60005b8d81101561050257600082828151811061040a5761040a610f0b565b60200260200101516000015184838151811061042857610428610f0b565b6020026020010151600001510190508a828151811061044957610449610f0b565b6020026020010151600001518111156104f957808b838151811061046f5761046f610f0b565b6020026020010151600001818152505084828151811061049157610491610f0b565b60200260200101518383815181106104ab576104ab610f0b565b6020026020010151602001516040516020016104c8929190610f3a565b6040516020818303038152906040528b83815181106104e9576104e9610f0b565b6020026020010151602001819052505b506001016103ee565b5050505050505b60005b88811015610b2f57808314610b275760008a8a8381811061052f5761052f610f0b565b90506040020160200160208101906105479190610f69565b73ffffffffffffffffffffffffffffffffffffffff168b8b8681811061056f5761056f610f0b565b90506040020160200160208101906105879190610f69565b73ffffffffffffffffffffffffffffffffffffffff16109050366000821561062d578a8a878181106105bb576105bb610f0b565b90506020028101906105cd9190610fa6565b90506000036105de57505050610b27565b8a8a878181106105f0576105f0610f0b565b90506020028101906106029190610fa6565b8581811061061257610612610f0b565b90506020028101906106249190610fa6565b915091506106ad565b8a8a8581811061063f5761063f610f0b565b90506020028101906106519190610fa6565b905060000361066257505050610b27565b8a8a8581811061067457610674610f0b565b90506020028101906106869190610fa6565b8781811061069657610696610f0b565b90506020028101906106a89190610fa6565b915091505b8015610b235760008060005b838110156108415760008060008787858060010196508181106106de576106de610f0b565b905060200201359050608081901c9250806fffffffffffffffffffffffffffffffff16915050600087878580600101965081811061071e5761071e610f0b565b9050602002013590508e5160001461075b576107558f8b8151811061074557610745610f0b565b6020026020010151848484610b58565b90935091505b6107848e8b8151811061077057610770610f0b565b602002602001015160200151848484610b58565b909350915088610792579091905b62ffffff60a082901c168b02600160ff60d884901c161461080d5760008989878060010198508181106107c7576107c7610f0b565b905060200201359050848b6107ee57816fffffffffffffffffffffffffffffffff166107f4565b608082901c5b03620f42400282111561080b5750505050506106b9565b505b8084620f42400201838202816108255761082561100e565b0490508581111561083857909550935084845b505050506106b9565b8115610b1f578f8f8881811061085957610859610f0b565b905060400201600001353a600260d886901c60ff1610610884576a0459480000000000000000610891565b6a0186a000000000000000005b6affffffffffffffffffffff1602816108ac576108ac61100e565b04820391508a87815181106108c3576108c3610f0b565b602002602001015160000151821315610b1f57818b88815181106108e9576108e9610f0b565b6020026020010151600001818152505060008b888151811061090d5761090d610f0b565b602002602001015160200151905060008c8b8151811061092f5761092f610f0b565b6020026020010151602001515160200190506000825190506000811115610975575b818110156109755782810160200151801561096c5750610975565b50602001610951565b80821115610a035760005b835181101561099757600084820152602001610980565b508167ffffffffffffffff8111156109b1576109b1610edc565b6040519080825280601f01601f1916602001820160405280156109db576020820181803683370190505b508e8b815181106109ee576109ee610f0b565b60200260200101516020018190529250610a48565b60405182602085010181811115610a1957806040525b508290505b8351811015610a3b57600081602086010152602081019050610a1e565b5081835114610a48578183525b5060008d8c81518110610a5d57610a5d610f0b565b60200260200101516020015190506000602090505b82811015610a8a578181015184820152602001610a72565b508a905060005b818265ffffffffffff1614610aac57600891821c9101610a91565b600882901b915080821791505060a081901b867f7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff161795508815610b10577f8000000000000000000000000000000000000000000000000000000000000000861795505b50018390526001871b99909917985b5050505b5050505b60010161050c565b505b505b600101610125565b5080600003610b4a5750610b4f565b610122565b50505050505050565b60008062ffffff60a084901c168360205b88518111610cc5578881015173ffffffffffffffffffffffffffffffffffffffff80821690841603610cb25765ffffffffffff60a882901c1660ff60a083901c161b6000610bb7868361106c565b9050600160ff60d88b901c16147f80000000000000000000000000000000000000000000000000000000000000008416600003610c505781610bfc8c620f424061106c565b610c069190611089565b610c108d8461106c565b610c1a919061109c565b610c24908d6110d7565b9b5080610c3d57610c38620f42408361109c565b610c3f565b825b610c49908c611089565b9a50610cae565b81610c5e8d620f424061106c565b610c689190611089565b610c728c8461106c565b610c7c919061109c565b610c86908c6110d7565b9a5080610c9f57610c9a620f42408361109c565b610ca1565b825b610cab908d611089565b9b505b5050505b50610cbe602082611089565b9050610b69565b50959794965093945050505050565b60008083601f840112610ce657600080fd5b50813567ffffffffffffffff811115610cfe57600080fd5b6020830191508360208260051b8501011115610d1957600080fd5b9250929050565b600080600080600080600060a0888a031215610d3b57600080fd5b873596506020880135955060408801359450606088013567ffffffffffffffff811115610d6757600080fd5b8801601f81018a13610d7857600080fd5b803567ffffffffffffffff811115610d8f57600080fd5b8a60208260061b8401011115610da457600080fd5b60209190910194509250608088013567ffffffffffffffff811115610dc857600080fd5b610dd48a828b01610cd4565b989b979a50959850939692959293505050565b60005b83811015610e02578181015183820152602001610dea565b50506000910152565b6000602082016020835280845180835260408501915060408160051b86010192506020860160005b82811015610ed0577fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc087860301845281518051865260208101519050604060208701528051806040880152610e8f816060890160208501610de7565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169590950160600194506020938401939190910190600101610e33565b50929695505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60008351610f4c818460208801610de7565b835190830190610f60818360208801610de7565b01949350505050565b600060208284031215610f7b57600080fd5b813573ffffffffffffffffffffffffffffffffffffffff81168114610f9f57600080fd5b9392505050565b60008083357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe1843603018112610fdb57600080fd5b83018035915067ffffffffffffffff821115610ff657600080fd5b6020019150600581901b3603821315610d1957600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b80820281158282048414176110835761108361103d565b92915050565b808201808211156110835761108361103d565b6000826110d2577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b500490565b818103818111156110835761108361103d56fea2646970667358221220daa3abe8b00d0b34e0199ca63271e0bc081a2ccb032eef4c67a3426f84ff239464736f6c637828302e382e32372d6e696768746c792e323032342e372e32352b636f6d6d69742e30363563326333640059")
