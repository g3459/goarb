contract Caller {

    uint internal constant STATE_MASK=0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint internal constant TYPE_MASK=0x00000000ff000000000000000000000000000000000000000000000000000000;
    address owner;
    
    constructor() payable{
        unchecked{owner=msg.sender;}
    }

    // function setAddress(address a,bool b) external payable check{
    //     unchecked{whitelist[a]=b;}
    // }

    fallback() external payable check{
        unchecked{
            for(uint i;i<msg.data.length;i+=32){
                assembly{
                    let poolCall:=calldataload(i)
                    let rstate:=and(poolCall,STATE_MASK)
                    if rstate{
                        switch and(poolCall,TYPE_MASK)
                        case 0x0000000001000000000000000000000000000000000000000000000000000000 {
                            mstore(0x80,0x0902f1ac00000000000000000000000000000000000000000000000000000000)
                        }case 0x0000000002000000000000000000000000000000000000000000000000000000 {
                            mstore(0x80,0xe76c01e400000000000000000000000000000000000000000000000000000000)
                        }default{
                            mstore(0x80,0x3850c7bd00000000000000000000000000000000000000000000000000000000)
                        }
                        pop(call(gas(), poolCall, 0, 0x80, 0x04, 0x80, 0x20))
                        if xor(and(keccak256(0x80,0x20),STATE_MASK),rstate){
                            revert(0,0)
                        }
                    }
                }
            }
            for(uint i;i<msg.data.length;i+=32){
                uint poolCall=uint(bytes32(msg.data[i:]));
                uint amIn=uint(uint48(poolCall>>168))<<uint8(poolCall>>160);
                address pool;
                assembly{pool:=poolCall}
                bool direc;
                assembly{direc:=and(poolCall,0x8000000000000000000000000000000000000000000000000000000000000000)}
                if(poolCall&TYPE_MASK==0x0000000001000000000000000000000000000000000000000000000000000000){
                    uint r0;uint r1;
                    assembly{
                        mstore(0x80,0x0902f1ac00000000000000000000000000000000000000000000000000000000)
                        pop(call(gas(), pool, 0, 0x80, 0x04, 0x80, 0x40))
                        r0:=mload(0x80)
                        r1:=mload(0xa0)
                    }
                    uint amOut=amIn*997;
                    amOut = (direc
                        ? (amOut * r1) / (r0 * 1000 + amOut)
                        : (amOut * r0) / (r1 * 1000 + amOut));
                    IERC20(direc?IUniV2Pool(pool).token0():IUniV2Pool(pool).token1()).transfer(pool,amIn);
                    IUniV2Pool(pool).swap(direc?0:amOut, direc?amOut:0, address(this), "");
                }else{
                    IUniV3Pool(pool).swap(address(this), direc, int(amIn) , direc ? 4295128740 : 1461446703485210103287273052203988822378723970341, "");
                }
            }
        }
    }

    receive() external payable{}

    function recover() external payable check{
        unchecked{payable(msg.sender).transfer(address(this).balance);}
    }

    function execute(address target, bytes calldata call) external payable check returns (bool s){
        unchecked{(s,)=target.call(call);}
    }

    modifier check{
        _;
        unchecked{require(owner==tx.origin || owner==msg.sender);}
    }

    // function uniswapV3FlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function flashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function algebraFlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    function uniswapV3SwapCallback(int am0 , int am1, bytes calldata) external payable check{
        unchecked{IERC20(am0>am1?IUniV3Pool(msg.sender).token0():IUniV3Pool(msg.sender).token1()).transfer(msg.sender,uint(am0>am1?am0:am1));}
    }

    // function swapCallback(int, int, bytes calldata data) external payable {unchecked{address(this).call(data);} }

    function algebraSwapCallback(int am0, int am1, bytes calldata) external payable check{
        unchecked{IERC20(am0>am1?IUniV3Pool(msg.sender).token0():IUniV3Pool(msg.sender).token1()).transfer(msg.sender,uint(am0>am1?am0:am1));}
    }

    // function apeCall(address, uint, uint, bytes calldata data) external payable {unchecked{address(this).call(data);} }

}

interface IERC20{
    function balanceOf(address ) external view returns ( uint256 );
    function transfer(address, uint256 ) external returns ( uint256 );
}


interface IUniV3Pool{
    function token0() external view returns ( address );
    function token1() external view returns ( address );
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
}

interface IUniV2Pool{
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
    function token0() external view returns ( address );
    function token1() external view returns ( address );
}

interface IAlgebraV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointIndex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}