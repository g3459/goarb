
contract Caller {

    bool lock;
    mapping(address=>bool) public whitelist;

    constructor() payable{
        unchecked{
            whitelist[msg.sender]=true;
        }
    }

    function setAddress(address a,bool b) external payable locked{
        unchecked{
            whitelist[a]=b;
        }
    }

    fallback() external payable{
        unchecked{
            bytes calldata _calls=msg.data[32:];
            while(_calls.length>0){
                (, bytes memory state)=address(bytes20(_calls)).staticcall(abi.encodeWithSelector(IUniV3Pool.slot0.selector));
                require(bytes4(keccak256(state))&0xfffffffe==bytes4(_calls[20:24])&0xfffffffe,"1");
                _calls=_calls[24:];
            }
            executeRoute(msg.data);
        }
    }

    function executeRoute(bytes calldata calls) internal locked{
        unchecked{
            uint amIn=uint256(uint128(bytes16(calls)));
            calls=calls[16:];
            uint gasPQ=uint256(uint128(bytes16(calls)));
            calls=calls[16:];
            while(calls.length>0){
                bool direc=bytes1(calls[23:24])&0x01==bytes1(0x01);
                amIn-=(amIn*85000)/gasPQ;
                (int am0,int am1)=IUniV3Pool(address(bytes20(calls))).swap(address(this), direc, int(amIn) , direc ? 4295128740 : 1461446703485210103287273052203988822378723970341, "");
                amIn=uint(-(am0>am1?am1:am0));
                calls=calls[24:];
            }
        }
    }

    function execute(address target, bytes calldata call) public payable locked returns (bool s){
        unchecked{
            (s,)=target.call(call);
        }
    }

    modifier locked{
        if(lock){
            _;
        }else{
            require(whitelist[msg.sender],"2");
            lock=true;
            _;
            lock=false;
        }
    }


    // function uniswapV3FlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function flashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function algebraFlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    function uniswapV3SwapCallback(int am0 , int am1, bytes calldata) external payable{
        unchecked{
            require(lock,"3");     
            address token=am0>am1?IUniV3Pool(msg.sender).token0():IUniV3Pool(msg.sender).token1();
            IERC20(token).transfer(msg.sender,uint(am0>am1?am0:am1));
        }
    }

    // function swapCallback(int, int, bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function algebraSwapCallback(int, int, bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function uniswapV2Call(address, uint, uint, bytes calldata data) external payable {unchecked{address(this).call(data);} }

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